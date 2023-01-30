package idefixgo

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

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
	Response string      `json:"r,omitempty" msgpack:"re,omitempty"`
	To       string      `json:"t" msgpack:"to"`
	Err      error       `json:"e,omitempty" msgpack:"er,omitempty"`
	Data     interface{} `json:"d,omitempty" msgpack:"dt,omitempty"`
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
	Address   string                 `bson:"address" json:"address" msgpack:"address"`
	Domain    string                 `bson:"domain" json:"domain" msgpack:"domain"`
	Timestamp time.Time              `bson:"timestamp" json:"timestamp" msgpack:"timestamp"`
	Meta      map[string]interface{} `bson:"meta" json:"meta" msgpack:"meta"`
	Schema    string                 `bson:"schema" json:"schema" msgpack:"schema"`
	Payload   interface{}            `bson:"payload" json:"payload" msgpack:"payload"`
}

func (e *Event) String() string {
	return fmt.Sprintf("[%s] %s @ %s | %s: %v | %v", e.Timestamp, e.Address, e.Domain, e.Schema, e.Payload, e.Meta)
}

type GetEventResponse struct {
	Events         []Event `json:"events" msgpack:"events"`
	ContinuationID string  `json:"cid" msgpack:"cid"`
}

type Schema struct {
	Description string `bson:"description" json:"description" msgpack:"description"`
	Hash        string `bson:"hash" json:"hash" msgpack:"hash"`
	Payload     string `bson:"payload" json:"payload" msgpack:"payload"`
}
