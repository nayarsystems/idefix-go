package messages

import (
	"fmt"
	"time"
)

/********************/
/*   Address Rules  */
/********************/

type MongoRulesExpression string

type AddressRules struct {
	// "Allow" rules to apply to every message reaching this address
	Allow MongoRulesExpression `json:"allow,omitempty"`

	// "Deny" rules to apply to every message reaching this address
	Deny MongoRulesExpression `json:"deny,omitempty"`
}

/*************/
/*  Domains  */
/*************/

type DomainInfo struct {
	// Domain name
	Domain string `bson:"_id" json:"domain" msgpack:"domain" mapstructure:"domain,omitempty"`

	// List of addresses which have admin permissions on this domain
	Admins []string `bson:"admins" json:"admins" msgpack:"admins" mapstructure:"admins"`

	// "Allow" rules to apply to every message reaching an address in this domain
	Allow string `bson:"allow" json:"allow" msgpack:"allow" mapstructure:"allow,omitempty"`

	// "Deny" rules to apply to every message reaching an address in this domain
	Deny string `bson:"deny" json:"deny" msgpack:"deny" mapstructure:"deny,omitempty"`

	// Variables added to the available environment during the rules execution
	Env map[string]interface{} `bson:"env" json:"env" msgpack:"env" mapstructure:"env"`
}

type Domain struct {
	DomainInfo `bson:",inline" mapstructure:",squash"`
	Creation   time.Time `bson:"creation" json:"creation" msgpack:"creation" mapstructure:"creation,omitempty"`
	LastUpdate time.Time `bson:"lastUpdate" json:"lastUpdate" msgpack:"lastUpdate" mapstructure:"lastUpdate,omitempty"`
}

/************/
/*  Events  */
/************/

type Event struct {
	EventMsg  `bson:",inline" mapstructure:",squash"`
	Domain    string    `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty"`
	Address   string    `json:"address" msgpack:"address" mapstructure:"address,omitempty"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp" msgpack:"timestamp" mapstructure:"timestamp,omitempty"`
}

func (e *Event) String() string {
	return fmt.Sprintf("[%s] %s @ %s | %s: %v | %v", e.Timestamp, e.Address, e.Domain, e.Type, e.Payload, e.Meta)
}

/********************/
/*  Bstates events  */
/********************/

type SchemaInfo struct {
	Address     string `bson:"address" json:"address" msgpack:"address"`
	Domain      string `bson:"domain" json:"domain" msgpack:"domain"`
	Description string `bson:"description" json:"description" msgpack:"description"`
	Hash        string `bson:"hash" json:"hash" msgpack:"hash"`
	Payload     string `bson:"payload" json:"payload" msgpack:"payload"`
}

type Schema struct {
	SchemaInfo   `bson:",inline" mapstructure:",squash"`
	CreationTime time.Time `bson:"creationTime" json:"creationTime" msgpack:"creationTime"`
}
