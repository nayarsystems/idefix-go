package messages

import (
	"time"

	"github.com/nayarsystems/mapstructure"
)

/************/
/*   Login  */
/************/

type LoginMsg struct {
	Address  string                 `json:"address" msgpack:"address" mapstructure:"address,omitempty"`
	Encoding string                 `json:"encoding" msgpack:"encoding" mapstructure:"encoding,omitempty"`
	Token    string                 `json:"token" msgpack:"token" mapstructure:"token,omitempty"`
	Time     int64                  `json:"time" msgpack:"time" mapstructure:"time,omitempty"`
	Meta     map[string]interface{} `json:"meta" msgpack:"meta" mapstructure:"meta,omitempty"`
}

/********************/
/*      Address     */
/********************/

type AddressTokenResetMsg struct {
	// Address to query
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty"`
}

type AddressDisableMsg struct {
	// Address to query
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty"`
	// Disable or enable the address
	Disabled bool `json:"disabled" msgpack:"disabled" mapstructure:"disabled,omitempty"`
}

type AddressAccessRulesGetMsg struct {
	// Address to query
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty"`
}

type AddressAccessRulesGetResponseMsg struct {
	AccessRules string `json:"accessRules" msgpack:"accessRules" mapstructure:"accessRules,omitempty"`
}

type AddressAccessRulesUpdateMsg struct {
	// Address to query
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty"`

	AccessRules string `json:"accessRules" msgpack:"accessRules" mapstructure:"accessRules,omitempty"`
}

type AddressAccessRulesUpdateResponseMsg struct {
}

type AddressDomainGetMsg struct {
	// Address to query
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty"`
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
	Since time.Time `json:"since" msgpack:"since" mapstructure:"since,omitempty"`

	// Limit the number of results returned
	Limit uint `json:"limit" msgpack:"limit" mapstructure:"limit,omitempty"`

	// Timeout sets the long-polling duration
	Timeout time.Duration `json:"timeout" msgpack:"timeout" mapstructure:"timeout,omitempty"`

	// ContinuationID lets you get following results after your last request
	ContinuationID string `json:"cid" msgpack:"cid" mapstructure:"cid,omitempty"`
}

func (m *EventsGetMsg) ToMsi() (data msi, err error) {
	data, err = ToMsiGeneric(m,
		mapstructure.ComposeEncodeFieldMapHookFunc(
			EncodeTimeToStringHook(time.RFC3339),
			EncodeDurationToStringHook()))

	return data, err
}

func (m *EventsGetMsg) ParseMsi(input msi) (err error) {
	err = ParseMsiGeneric(input, m,
		mapstructure.ComposeDecodeHookFunc(
			DecodeAnyTimeStringToTimeHookFunc(),
			mapstructure.StringToTimeDurationHookFunc(),
		))
	if err != nil {
		return err
	}
	return nil
}

type EventsGetUIDResponseMsg struct {
	Event `bson:",inline" mapstructure:",squash"`
}

type EventsGetResponseMsg struct {
	// Array of events
	Events []*Event `json:"events" msgpack:"events" mapstructure:"events,omitempty"`

	// ContinuationID lets you get following results after your last request
	ContinuationID string `json:"cid" msgpack:"cid" mapstructure:"cid,omitempty"`
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
	Domain `bson:",inline" mapstructure:",squash"`
}

type DomainCreateResponseMsg struct {
	// Domain name
	Domain `bson:",inline" mapstructure:",squash"`
}

type DomainUpdateMsg struct {
	Domain `bson:",inline" mapstructure:",squash"`
}

type DomainUpdateResponseMsg struct {
	Domain `bson:",inline" mapstructure:",squash"`
}

type DomainUpdateAccessRulesMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty"`

	// Access rules (javascript snippet by default) to be applied to every message reaching an address in this domain
	AccessRules string `json:"accessRules" msgpack:"accessRules" mapstructure:"accessRules,omitempty"`
}

type DomainUpdateAccessRulesResponseMsg struct {
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

/*************/
/*  Groups  */
/*************/

type GroupAddAddressMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain"`

	// Group name
	Group string `json:"group" msgpack:"group" mapstructure:"group"`

	// Address to assign
	Address string `json:"address" msgpack:"address" mapstructure:"address"`
}
type GroupAddAddressResponseMsg struct {
}

type GroupRemoveAddressMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain"`

	// Group name
	Group string `json:"group" msgpack:"group" mapstructure:"group"`

	// Address to remove
	Address string `json:"address" msgpack:"address" mapstructure:"address"`
}

type GroupRemoveAddressResponseMsg struct {
}

type GroupGetAddressesMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain"`

	// Group name
	Group string `json:"group" msgpack:"group" mapstructure:"group"`
}

type GroupGetAddressesResponseMsg struct {
	// Addresses in group: domain -> addresses
	Addresses map[string][]string `json:"addresses" msgpack:"addresses" mapstructure:"addresses"`
}

type DomainGetGroupsMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain"`
}

type DomainGetGroupsResponseMsg struct {
	// Groups <domain>#<group>
	Groups []string `json:"groups" msgpack:"groups" mapstructure:"groups"`
}

type AddressGetGroupsMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty"`

	// Address
	Address string `json:"address" msgpack:"address" mapstructure:"address"`
}

type AddressGetGroupsResponseMsg struct {
	// Address's groups: list of domain#group
	Groups []string `json:"groups" msgpack:"groups" mapstructure:"groups"`
}
