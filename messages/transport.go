package messages

type Message struct {
	Res  string      `json:"r,omitempty" msgpack:"re,omitempty"`
	To   string      `json:"t" msgpack:"to"`
	Err  string      `json:"e,omitempty" msgpack:"er,omitempty"`
	Data interface{} `json:"d,omitempty" msgpack:"dt,omitempty"`
}
