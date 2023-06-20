package messages

import (
	"fmt"
	"time"

	"github.com/jaracil/ei"
)

/************/
/*   Login  */
/************/

type LoginMsg struct {
	Address  string                 `json:"address" msgpack:"address"`
	Encoding string                 `json:"encoding" msgpack:"encoding"`
	Token    string                 `json:"token" msgpack:"token"`
	Time     int64                  `json:"time" msgpack:"time"`
	Meta     map[string]interface{} `json:"meta" msgpack:"meta"`
}

/********************/
/*      Address     */
/********************/

type AddressTokenResetMsg struct {
	// Address to query
	Address string `bson:"address" json:"address" msgpack:"address" mapstructure:"address,omitempty"`
}

type AddressDisableMsg struct {
	// Address to query
	Address string `bson:"address" json:"address" msgpack:"address" mapstructure:"address,omitempty"`
	// Disable or enable the address
	Disabled bool `bson:"disabled" json:"disabled" msgpack:"disabled" mapstructure:"disabled,omitempty"`
}

type AddressRulesGetMsg struct {
	// Address to query
	Address string `bson:"address" json:"address" msgpack:"address" mapstructure:"address,omitempty"`
}

type AddressRulesUpdateMsg struct {
	// Address to query
	Address string `bson:"address" json:"address" msgpack:"address" mapstructure:"address,omitempty"`

	AddressRules `bson:",inline" mapstructure:",squash"`
}

/************/
/*  Events  */
/************/

type EventMsg struct {
	// UID must be provided by the client, and must be a unique identifier
	UID string `bson:"uid" json:"uid" msgpack:"uid" mapstructure:"uid,omitempty"`

	// Meta can hold any client provided data related to this event
	Meta map[string]interface{} `bson:"meta" json:"meta" msgpack:"meta" mapstructure:"meta,omitempty"`

	// Type parameter holds a mimetype or similar identifier of the payload
	Type string `bson:"type" json:"type" msgpack:"type" mapstructure:"type,omitempty"`

	// Payload is the raw data of the event
	Payload interface{} `bson:"payload" json:"payload" msgpack:"payload" mapstructure:"payload,omitempty"`
}

type EventResponseMsg struct {
	Ok     bool `bson:"ok" json:"ok" msgpack:"ok" mapstructure:"ok"`
	Schema bool `bson:"schema" json:"schema" msgpack:"schema" mapstructure:"schema"`
}

type EventsGetMsg struct {
	// UID of the event
	UID string `bson:"uid" json:"uid" msgpack:"uid" mapstructure:"uid,omitempty"`

	// Domain name to get all events from
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty"`

	// Address of the event creator
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty"`

	// Timestamp to search since
	Since time.Time `json:"since" msgpack:"since" mapstructure:"-,omitempty"`

	// Limit the number of results returned
	Limit uint `json:"limit" msgpack:"limit" mapstructure:"limit,omitempty"`

	// Timeout sets the long-polling duration
	Timeout time.Duration `json:"timeout" msgpack:"timeout" mapstructure:"-,omitempty"`

	// ContinuationID lets you get following results after your last request
	ContinuationID string `json:"cid" msgpack:"cid" mapstructure:"cid,omitempty"`
}

func (m *EventsGetMsg) ToMsi() (data msi, err error) {
	data, err = ToMsiGeneric(m, nil)
	if err != nil {
		return nil, err
	}
	// replace timeout field by its string format
	data["timeout"] = m.Timeout.String()
	data["since"] = m.Since.Format(time.RFC3339)

	return data, err
}

func (m *EventsGetMsg) ParseMsi(input msi) (err error) {
	m.Since, _ = ei.N(input).M("since").Time()
	toutraw, err := ei.N(input).M("timeout").Raw()
	var tout time.Duration
	if err == nil {
		toutStr, isStr := toutraw.(string)
		if !isStr {
			return fmt.Errorf("'timeout' field mut be a string formatted duration")
		}
		tout, err = time.ParseDuration(toutStr)
		if err != nil {
			return fmt.Errorf("can't parse 'timeout' field: %v", err)
		}
		delete(input, "timeout")
	}
	err = ParseMsiGeneric(input, m, nil)
	if err != nil {
		return err
	}
	m.Timeout = tout
	return nil
}

type EventsGetUIDResponseMsg struct {
	Event `bson:",inline" mapstructure:",squash"`
}

func (m *EventsGetUIDResponseMsg) ToMsi() (data msi, err error) {
	return m.Event.ToMsi()
}

func (m *EventsGetUIDResponseMsg) ParseMsi(input msi) (err error) {
	return m.Event.ParseMsi(input)
}

type EventsGetResponseMsg struct {
	// Array of events
	Events []*Event `json:"events" msgpack:"events"`

	// ContinuationID lets you get following results after your last request
	ContinuationID string `json:"cid" msgpack:"cid"`
}

func (m *EventsGetResponseMsg) ToMsi() (data msi, err error) {
	data = msi{
		"cid": m.ContinuationID,
	}
	events := []msi{}
	for _, ev := range m.Events {
		evraw, err := ev.ToMsi()
		if err != nil {
			return nil, err
		}
		events = append(events, evraw)
	}
	data["events"] = events
	return data, err
}

func (m *EventsGetResponseMsg) ParseMsi(input msi) (err error) {
	m.ContinuationID, err = ei.N(input).M("cid").String()
	if err != nil {
		return err
	}
	revents, err := ei.N(input).M("events").Slice()
	if err != nil {
		return err
	}
	for _, rev := range revents {
		ev := &Event{}
		err = ParseMsg(rev, ev)
		if err != nil {
			return err
		}
		m.Events = append(m.Events, ev)
	}
	return nil
}

/********************/
/*  Bstates events  */
/********************/

type SchemaMsg struct {
	// A human readable description of the schema
	Description string `bson:"description" json:"description" msgpack:"description" mapstructure:"description,omitempty"`

	// Schema content
	Payload string `bson:"payload,omitempty" json:"payload,omitempty" msgpack:"payload,omitempty" mapstructure:"payload,omitempty"`
}

type SchemaResponseMsg struct {
	SchemaMsg `bson:",inline" mapstructure:",squash"`
	Hash      string `bson:"hash" json:"hash" msgpack:"hash" mapstructure:"hash,omitempty"`
}

type SchemaGetMsg struct {
	// Hash of the schema requested
	Hash string `bson:"hash" json:"hash" msgpack:"hash" mapstructure:"hash,omitempty"`

	// Check if the schema is available, but do not return its content
	Check bool `bson:"check,omitempty" json:"check,omitempty" msgpack:"check,omitempty" mapstructure:"check,omitempty"`
}

type SchemaGetResponseMsg struct {
	SchemaMsg `bson:",inline" mapstructure:",squash"`

	// Hash of the schema requested
	Hash string `bson:"hash" json:"hash" msgpack:"hash" mapstructure:"hash,omitempty"`
}

/*************/
/*  Domains  */
/*************/

type DomainGetMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty"`
}

type DomainDeleteMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty"`
}

type DomainCreateMsg struct {
	// Domain name
	DomainInfo `bson:",inline" mapstructure:",squash"`
}

type DomainCreateResponseMsg struct {
	// Domain name
	Domain `bson:",inline" mapstructure:",squash"`
}

type DomainUpdateMsg struct {
	DomainInfo `bson:",inline" mapstructure:",squash"`
}

type DomainUpdateResponseMsg struct {
	Domain `bson:",inline" mapstructure:",squash"`
}

type DomainAssignMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty"`

	// Address to assign
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty"`
}

type DomainGetTreeMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty"`
}

// TODO: transform to struct
type DomainGetTreeResponseMsg []string

type DomainCountAddressesMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty"`
}

type DomainCountAddressesResponseMsg struct {
	Addresses int `json:"addresses" msgpack:"addresses" mapstructure:"addresses,omitempty"`
}

type DomainListAddressesMsg struct {
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty"`
	Limit  uint   `json:"limit" msgpack:"limit" mapstructure:"limit,omitempty"`
	Skip   uint   `json:"skip" msgpack:"skip" mapstructure:"skip,omitempty"`
}

type DomainListAddressesResponseMsg struct {
	Addresses map[string]string `json:"addresses" msgpack:"addresses" mapstructure:"addresses,omitempty"`
}
