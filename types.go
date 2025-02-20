package idefixgo

import (
	"github.com/spf13/viper"
)

// ClientOptions defines the configuration options for initializing a Client.
//
// These options include connection details, security settings, metadata, and other parameters
// that influence how the Client interacts with the MQTT broker
type ClientOptions struct {
	Broker    string                 `json:"broker"`            // The address or URL of the MQTT broker the client will connect to.
	Encoding  string                 `json:"encoding"`          // Specifies the data encoding format to be used.
	CACert    []byte                 `json:"cacert,omitempty"`  // A byte slice containing the Certificate Authority (CA) certificate for secure communication.
	Address   string                 `json:"address"`           // The specific client address or identifier used for communications.
	Token     string                 `json:"token"`             // A security token for authenticating the client to the broker.
	Meta      map[string]interface{} `json:"meta,omitempty"`    // A map containing additional metadata.
	SessionID string                 `json:"session,omitempty"` // A string representing the session ID for the client's connection, useful for session management.
	Groups    []string               `json:"groups,omitempty"`  // A list of group identifiers to which the client belongs, used for access control.
	vp        *viper.Viper
}
