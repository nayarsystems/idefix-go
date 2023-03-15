package idefixgo

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	. "github.com/nayarsystems/idefix-go/errors"
	"github.com/nayarsystems/idefix/core/idefix/normalize"
	"github.com/vmihailenco/msgpack/v5"
)

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
		return ErrMarshal
	}

	if !marshaled {
		return ErrMarshal.With("unsupported encoding")
	}

	var compressed bool

	if strings.Contains(c.opts.Encoding, "g") && !compressed { // TODO: Compression threshold?
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
	if msg.Error() != nil {
		return ErrInternal.With(msg.Error().Error())
	}
	return nil
}

func (c *Client) receiveMessage(client mqtt.Client, msg mqtt.Message) {
	if !strings.HasPrefix(msg.Topic(), c.responseTopic()) {
		return
	}

	topicChuncks := strings.Split(msg.Topic(), "/")
	if len(topicChuncks) != 4 {
		return
	}

	flags := topicChuncks[3]
	payload := msg.Payload()

	var tm Message
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
			// fmt.Printf("can't decompress gzip: %v\n", err)
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
		// fmt.Printf("unmarshal error decoding message: %v\n", unmarshalErr)
		return
	}

	if !unmarshaled {
		// fmt.Println("unmarshal error: codec not found ")
		return
	}

	if strings.HasPrefix(tm.To, c.opts.Address+".") {
		return
	}

	tm.To = strings.TrimPrefix(tm.To, c.opts.Address+".")

	if tm.To == "" {
		return
	}

	if msiData, ok := tm.Data.(map[string]interface{}); ok {
		err := normalize.DecodeTypes(msiData)
		if err != nil {
			// fmt.Println("error decoding mqtt message", err)
			return
		}
	}

	if n := c.ps.Publish(tm.To, &tm); n == 0 {
		// fmt.Printf("mqtt message published but there is no receivers: %#v\n", tm)
	}
}
