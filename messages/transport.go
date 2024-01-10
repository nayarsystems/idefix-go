package messages

// Fix: json marshal and msgpack marshal do not share the same field names
type Message struct {
	To   string      `json:"t" msgpack:"to"`
	Data interface{} `json:"d" msgpack:"dt"` // Cannot omitempty because the zero value is a valid value
	Res  string      `json:"r,omitempty" msgpack:"re,omitempty"`
	Err  string      `json:"e,omitempty" msgpack:"er,omitempty"`
}
