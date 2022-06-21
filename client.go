package idefixgo

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/vmihailenco/msgpack/v5"
	"gitlab.com/garagemakers/idefix/core/idefix"
	"gitlab.com/garagemakers/idefix/core/idefix/normalize"
	"gitlab.com/garagemakers/idefix/modules/transport/transportMqtt"
)

func NewClient(pctx context.Context, opts *ConnectionOptions) (*Client, error) {
	im := &Client{
		opts:     opts,
		Encoding: opts.Encoding,
	}

	if err := im.connect(pctx, opts.BrokerAddress, opts.CACert); err != nil {
		return nil, err
	}

	if err := im.login(opts.Address, opts.Token, opts.Meta); err != nil {
		return nil, err
	}

	time.Sleep(time.Second * 3)
	return im, nil
}

func (im *Client) connect(pctx context.Context, brokerAddress string, CACert []byte) (err error) {
	im.ctx, im.cancelFunc = context.WithCancel(pctx)
	im.prefix = MqttIdefixPrefix
	im.Messages = make(chan *Message, 100)
	im.compThreshold = 128

	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerAddress)
	opts.SetCleanSession(true)
	opts.SetUsername("device")
	opts.SetPassword("77dev22p1")

	if len(CACert) > 0 {
		certpool := x509.NewCertPool()
		certpool.AppendCertsFromPEM(CACert)

		opts.SetTLSConfig(&tls.Config{
			RootCAs: certpool,
		})
	}

	im.sessionID, err = randSessionID()
	if err != nil {
		return err
	}
	opts.SetClientID(im.sessionID)

	opts.SetConnectionLostHandler(im.connectionLostHandler)
	opts.SetDefaultPublishHandler(im.messageHandler)

	im.client = mqtt.NewClient(opts)

	token := im.client.Connect()
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	token = im.client.Subscribe(fmt.Sprintf("%s/%s/r/+", im.prefix, im.sessionID), 1, nil)
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	fmt.Printf("%#v\n", im)

	return nil
}

func (im *Client) responseTopic() string {
	return fmt.Sprintf("%s/%s/r/", im.prefix, im.sessionID)
}

func (im *Client) publishTopic(flags string) string {
	return fmt.Sprintf("%s/%s/t/%s", im.prefix, im.sessionID, flags)
}

func (im *Client) login(deviceAddress string, deviceToken string, meta map[string]interface{}) (err error) {
	im.address = deviceAddress
	im.token = deviceToken

	lm := loginMsg{
		Address:  deviceAddress,
		Token:    deviceToken,
		Encoding: im.Encoding,
		Time:     time.Now().UnixMilli(),
		Meta:     meta,
	}

	tm := &Message{
		To:       TopicTransportLogin,
		Response: im.responseTopic(),
		Data:     lm,
	}

	return im.sendMessage(tm)
}

func randSessionID() (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("can't create connection ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (im *Client) sendMessage(tm *Message) (err error) {
	var flags string
	var marshaled bool
	var marshalErr error
	var data []byte

	if strings.Contains(im.Encoding, "j") && !marshaled {
		marshaled = true
		flags += "j"
		if v, ok := tm.Data.(map[string]interface{}); ok {
			opts := normalize.EncodeTypesOpts{BytesToB64: true}
			marshalErr = normalize.EncodeTypes(v, &opts)
		}
		if marshalErr == nil {
			data, marshalErr = json.Marshal(tm)
		}
	}

	if strings.Contains(im.Encoding, "m") && !marshaled {
		marshaled = true
		flags += "m"
		if v, ok := tm.Data.(map[string]interface{}); ok {
			opts := normalize.EncodeTypesOpts{}
			marshalErr = normalize.EncodeTypes(v, &opts)
		}
		if marshalErr == nil {
			data, marshalErr = msgpack.Marshal(tm)
		}
	}

	if marshalErr != nil {
		return idefix.ErrMarshall
	}

	if !marshaled {
		return fmt.Errorf("unsupported encoding")
	}

	var compressed bool

	if strings.Contains(im.Encoding, "g") && !compressed {
		buf := new(bytes.Buffer)
		wr := gzip.NewWriter(buf)
		n, err := wr.Write(data)
		wr.Close()
		if err == nil && n == len(data) {
			comp, err := io.ReadAll(buf)
			if err == nil && len(comp) < len(data) {
				compressed = true
				flags += "g"
				data = comp
			}
		}
	}

	msg := im.client.Publish(im.publishTopic(flags), 1, false, data)
	msg.Wait()
	return msg.Error()
}

func (im *Client) messageHandler(client mqtt.Client, msg mqtt.Message) {
	if !strings.HasPrefix(msg.Topic(), im.responseTopic()) {
		idefix.Warnf(im.ctx, "received alien message with topic: %s", msg.Topic())
		return
	}

	topicChuncks := strings.Split(msg.Topic(), "/")
	if len(topicChuncks) != 4 {
		idefix.Warnf(im.ctx, "received invalid topic: %s", msg.Topic())
		return
	}

	flags := topicChuncks[3]
	payload := msg.Payload()

	var tm transportMqtt.TransportMsg
	var unmarshalErr error
	var unmarshaled bool

	if strings.Contains(flags, "g") {
		r := bytes.NewReader(payload)
		gzr, err := gzip.NewReader(r)
		if err == nil {
			payload, err = io.ReadAll(gzr)
			gzr.Close()
		}
		if err != nil {
			idefix.Warnf(im.ctx, "can't decompress gzip: %v", err)
			return
		}
	}

	if strings.Contains(flags, "j") && !unmarshaled {
		unmarshalErr = json.Unmarshal(payload, &tm)
		unmarshaled = true
	}

	if strings.Contains(flags, "m") && !unmarshaled {
		unmarshalErr = msgpack.Unmarshal(payload, &tm)
		unmarshaled = true
	}

	if unmarshalErr != nil {
		idefix.Warnf(im.ctx, "unmarshal error decoding message: %v", unmarshalErr)
		return
	}

	if !unmarshaled {
		idefix.Warnf(im.ctx, "unmarshal error: codec not found ")
		return
	}

	if strings.HasPrefix(tm.To, im.address+".") {
		return
	}

	tm.To = strings.TrimPrefix(tm.To, im.address+".")

	if tm.To == "" {
		return
	}

	if msiData, ok := tm.Data.(map[string]interface{}); ok {
		err := normalize.DecodeTypes(msiData)
		if err != nil {
			// idefix.Warnf(im.ctx, "error decoding message types %s")
			return
		}
	}

	im.Messages <- &Message{Response: tm.Res, To: tm.To, Data: tm.Data, Err: tm.Err}
}

func (im *Client) connectionLostHandler(client mqtt.Client, err error) {
	im.cancelFunc()
}
