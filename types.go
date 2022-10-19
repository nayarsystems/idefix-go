package idefixgo

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type transportMsg struct {
	Res         string      `json:"r,omitempty" msgpack:"re,omitempty"`
	To          string      `json:"t" msgpack:"to"`
	Err         error       `json:"e,omitempty" msgpack:"er,omitempty"`
	Data        interface{} `json:"d,omitempty" msgpack:"dt,omitempty"`
	SrcSession  string
	SrcAddress  string
	SrcEncoding string
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
	Err      error       `json:"er" msgpack:"er"`
}

type ClientOptions struct {
	Broker   string                 `json:"broker"`
	Encoding string                 `json:"encoding"`
	CACert   []byte                 `json:"cacert,omitempty"`
	Address  string                 `json:"address"`
	Token    string                 `json:"token"`
	Meta     map[string]interface{} `json:"meta,omitempty"`

	vp *viper.Viper
}

type Event struct {
	Address   string                 `bson:"address" json:"address"`
	Domain    string                 `bson:"domain" json:"domain"`
	Timestamp time.Time              `bson:"timestamp" json:"timestamp"`
	Meta      map[string]interface{} `bson:"meta" json:"meta"`
	Schema    string                 `bson:"schema" json:"schema"`
	Payload   interface{}            `bson:"payload" json:"payload"`
}

func (e *Event) String() string {
	return fmt.Sprintf("[%s] %s @ %s | %s: %v | %v", e.Timestamp, e.Address, e.Domain, e.Schema, e.Payload, e.Meta)
}
