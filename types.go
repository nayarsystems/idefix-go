package idefixgo

import (
	"context"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	opts          *ConnectionOptions
	Encoding      string
	compThreshold int
	address       string
	token         string
	prefix        string
	client        mqtt.Client
	sessionID     string
	ctx           context.Context
	cancelFunc    context.CancelFunc

	Messages chan *Message
}

type loginMsg struct {
	Address  string                 `json:"address" msgpack:"address"`
	Encoding string                 `json:"encoding" msgpack:"encoding"`
	Token    string                 `json:"token" msgpack:"token"`
	Time     int64                  `json:"time" msgpack:"time"`
	Meta     map[string]interface{} `json:"meta" msgpack:"meta"`
}

// TopicTransportLogin is an absolute cloud path for login
const TopicTransportLogin = "idefix.login"

// Mqtt prefix for idefix project
const MqttIdefixPrefix = "ifx"

///

type Message struct {
	To       string      `json:"to" msgpack:"to"`
	Response string      `json:"re" msgpack:"re"`
	Data     interface{} `json:"dt" msgpack:"dt"`
	Err      string      `json:"er" msgpack:"er"`
}

type ConnectionOptions struct {
	BrokerAddress string
	Encoding      string
	CACert        []byte
	Address       string
	Token         string
	Meta          map[string]interface{}
}
