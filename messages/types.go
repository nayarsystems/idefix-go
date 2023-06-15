package messages

import (
	"fmt"
	"time"

	"github.com/jaracil/ei"
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
	Creation   time.Time `bson:"creation" json:"creation" msgpack:"creation" mapstructure:"-,omitempty"`
	LastUpdate time.Time `bson:"lastUpdate" json:"lastUpdate" msgpack:"lastUpdate" mapstructure:"-,omitempty"`
}

func (m *Domain) ToMsi() (data msi, err error) {
	data, err = ToMsiGeneric(m)
	if err != nil {
		return nil, err
	}
	data["creation"] = TimeToString(m.Creation)
	data["lastUpdate"] = TimeToString(m.LastUpdate)
	return data, err
}

func (m *Domain) ParseMsi(input msi) (err error) {
	err = ParseMsiGeneric(input, m)
	if err != nil {
		return err
	}
	m.LastUpdate, err = ei.N(input).M("lastUpdate").Time()
	if err != nil {
		return err
	}
	m.Creation, err = ei.N(input).M("creation").Time()
	if err != nil {
		return err
	}
	return nil
}

/************/
/*  Events  */
/************/

type Event struct {
	EventMsg  `bson:",inline" mapstructure:",squash"`
	Domain    string    `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty"`
	Address   string    `json:"address" msgpack:"address" mapstructure:"address,omitempty"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp" msgpack:"timestamp" mapstructure:"-,omitempty"`
}

func (m *Event) ToMsi() (data msi, err error) {
	data, err = ToMsiGeneric(m)
	if err != nil {
		return nil, err
	}
	// replace timeout field by its string format
	data["timestamp"] = TimeToString(m.Timestamp)
	return data, err
}

func (m *Event) ParseMsi(input msi) (err error) {
	m.Timestamp, _ = ei.N(input).M("timestamp").Time()
	err = ParseMsiGeneric(input, m)
	if err != nil {
		return err
	}
	return nil
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

func (m *Schema) ToMsi() (data msi, err error) {
	data, err = ToMsiGeneric(m)
	if err != nil {
		return nil, err
	}
	data["creationTime"] = TimeToString(m.CreationTime)
	return data, err
}

func (m *Schema) ParseMsi(input msi) (err error) {
	err = ParseMsiGeneric(input, m)
	if err != nil {
		return err
	}
	m.CreationTime, err = ei.N(input).M("creationTime").Time()
	if err != nil {
		return err
	}
	return nil
}

/*********************/
/*   Binary Updates  */
/*********************/

type TargetExec = int

const (
	LauncherTargetExec TargetExec = iota
	IdefixTargetExec
)
