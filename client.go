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
	"gitlab.com/garagemakers/idefix-go/minips"
	"gitlab.com/garagemakers/idefix/core/idefix/normalize"
)

func NewClient(pctx context.Context, opts *ConnectionOptions) (*Client, error) {
	c := &Client{
		opts: opts,
	}

	if err := c.connect(pctx, opts.BrokerAddress, opts.CACert); err != nil {
		return nil, err
	}

	if err := c.login(opts.Address, opts.Token, opts.Meta); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) connect(pctx context.Context, brokerAddress string, CACert []byte) (err error) {
	c.ctx, c.cancelFunc = context.WithCancel(pctx)
	c.prefix = MqttIdefixPrefix
	c.ps = minips.NewMinips[*Message](c.ctx)
	c.compThreshold = 128

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

	c.sessionID, err = randSessionID()
	if err != nil {
		return err
	}
	opts.SetClientID(c.sessionID)

	opts.SetConnectionLostHandler(c.connectionLostHandler)
	opts.SetDefaultPublishHandler(c.messageHandler)

	c.client = mqtt.NewClient(opts)

	token := c.client.Connect()
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	token = c.client.Subscribe(fmt.Sprintf("%s/%s/r/+", c.prefix, c.sessionID), 1, nil)
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (c *Client) responseTopic() string {
	return fmt.Sprintf("%s/%s/r/", c.prefix, c.sessionID)
}

func (c *Client) publishTopic(flags string) string {
	return fmt.Sprintf("%s/%s/t/%s", c.prefix, c.sessionID, flags)
}

func (c *Client) login(deviceAddress string, deviceToken string, meta map[string]interface{}) (err error) {
	c.address = deviceAddress
	c.token = deviceToken

	lm := loginMsg{
		Address:  c.opts.Address,
		Token:    c.opts.Token,
		Encoding: c.opts.Encoding,
		Meta:     c.opts.Meta,
		Time:     time.Now().UnixMilli(),
	}

	tm := &Message{
		To:   "login",
		Data: lm,
	}

	_, err = c.Call("idefix", tm, time.Second*3)
	if err != nil {
		return err
	}

	return nil
}

func randSessionID() (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("can't create connection ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (c *Client) sendMessage(tm *Message) (err error) {
	var flags string
	var marshaled bool
	var marshalErr error
	var data []byte

	if strings.Contains(c.opts.Encoding, "j") && !marshaled {
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

	if strings.Contains(c.opts.Encoding, "m") && !marshaled {
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
		return ErrMarshall
	}

	if !marshaled {
		return fmt.Errorf("unsupported encoding")
	}

	var compressed bool

	if strings.Contains(c.opts.Encoding, "g") && !compressed {
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

	msg := c.client.Publish(c.publishTopic(flags), 1, false, data)
	msg.Wait()
	return msg.Error()
}

func (c *Client) messageHandler(client mqtt.Client, msg mqtt.Message) {
	if !strings.HasPrefix(msg.Topic(), c.responseTopic()) {
		return
	}

	topicChuncks := strings.Split(msg.Topic(), "/")
	if len(topicChuncks) != 4 {
		return
	}

	flags := topicChuncks[3]
	payload := msg.Payload()

	var tm transportMsg
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
			fmt.Printf("can't decompress gzip: %v\n", err)
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
		fmt.Printf("unmarshal error decoding message: %v\n", unmarshalErr)
		return
	}

	if !unmarshaled {
		fmt.Println("unmarshal error: codec not found ")
		return
	}

	if strings.HasPrefix(tm.To, c.address+".") {
		return
	}

	tm.To = strings.TrimPrefix(tm.To, c.address+".")

	if tm.To == "" {
		return
	}

	if msiData, ok := tm.Data.(map[string]interface{}); ok {
		err := normalize.DecodeTypes(msiData)
		if err != nil {
			// fmt.Printf(im.ctx, "error decoding message types %s")
			return
		}
	}

	if n := c.ps.Publish(tm.To, &Message{Response: tm.Res, To: tm.To, Data: tm.Data, Err: tm.Err}); n == 0 {
		fmt.Println("Lost message:", tm.To)
	}
}

func (c *Client) connectionLostHandler(client mqtt.Client, err error) {
	c.cancelFunc()
}
