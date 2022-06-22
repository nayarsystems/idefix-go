package idefixgo

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
	BrokerAddress string
	Encoding      string
	CACert        []byte
	Address       string
	Token         string
	Meta          map[string]interface{}
}
