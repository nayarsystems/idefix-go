package idefixgo

import (
	"github.com/spf13/viper"
)

type ClientOptions struct {
	Broker   string                 `json:"broker"`
	Encoding string                 `json:"encoding"`
	CACert   []byte                 `json:"cacert,omitempty"`
	Address  string                 `json:"address"`
	Token    string                 `json:"token"`
	Meta     map[string]interface{} `json:"meta,omitempty"`

	vp *viper.Viper
}
