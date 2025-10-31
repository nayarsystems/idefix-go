package messages

import (
	"time"

	"github.com/jaracil/ei"
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
	Groups   []string               `json:"groups" msgpack:"groups" mapstructure:"groups,omitempty"`
	NoCreate bool                   `json:"no-create" msgpack:"no-create" mapstructure:"no-create,omitempty"`
}

/********************/
/*      Address     */
/********************/

type AddressTokenResetMsg struct {
	// Address to query
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty" validate:"required"`
}

type AddressDisableMsg struct {
	// Address to query
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty" validate:"required"`
	// Disable or enable the address
	Disabled bool `json:"disabled" msgpack:"disabled" mapstructure:"disabled,omitempty" validate:"required"`
}

type AddressAccessRulesGetMsg struct {
	// Address to query
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty" validate:"required"`
}

type AddressAccessRulesGetResponseMsg struct {
	AccessRules string `json:"accessRules" msgpack:"accessRules" mapstructure:"accessRules,omitempty"`
}

type AddressAccessRulesUpdateMsg struct {
	// Address to query
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty" validate:"required"`

	AccessRules string `json:"accessRules" msgpack:"accessRules" mapstructure:"accessRules,omitempty"`
}

type AddressAccessRulesUpdateResponseMsg struct {
}

type AddressDomainGetMsg struct {
	// Address to query
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty"`
}

type AddressStatesGetMsg struct {
	// Address to query
	Address string `json:"address" msgpack:"address" mapstructure:"address"`
}

type AddressStatesGetResMsg struct {
	States map[string]*StateEntry `json:"states" msgpack:"states" mapstructure:"states"`
}

func (m *AddressStatesGetResMsg) ParseMsi(input msi) (err error) {
	statesRaw, err := ei.N(input).M("states").MapStr()
	if err != nil {
		return err
	}
	m.States = make(map[string]*StateEntry)
	for k, v := range statesRaw {
		var se StateEntry
		err = ParseMsg(v, &se)
		if err != nil {
			return err
		}
		m.States[k] = &se

	}
	return nil
}

type AddressConfigGetMsg struct {
	// Address to query
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty"`
}

type AddressConfigGetResponseMsg struct {
	// Address' config
	Config []byte `json:"config" msgpack:"config" mapstructure:"config,omitempty"`
}

type AddressConfigUpdateMsg struct {
	// Address to query
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty"`

	// Config to update
	Config []byte `json:"config" msgpack:"config" mapstructure:"config,omitempty"`
}

type AddressConfigUpdateResponseMsg struct {
}

type AddressAliasGetMsg struct {
	// Address to query
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty"`
}

type AddressAliasGetResponseMsg struct {
	Alias []string `json:"alias" msgpack:"alias" mapstructure:"alias,omitempty"`
}

type AddressAliasAddMsg struct {
	// Address to query
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty"`

	// Alias to add
	Alias string `json:"alias" msgpack:"alias" mapstructure:"alias,omitempty"`
}

type AddressAliasAddResponseMsg struct {
	// Alias to add
	Alias []string `json:"alias" msgpack:"alias" mapstructure:"alias,omitempty"`
}

type AddressAliasRemoveMsg struct {
	// Address to query
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty"`

	// Alias to remove
	Alias string `json:"alias" msgpack:"alias" mapstructure:"alias,omitempty"`
}

type AddressAliasRemoveResponseMsg struct {
	// Alias to remove
	Alias []string `json:"alias" msgpack:"alias" mapstructure:"alias,omitempty"`
}

type AddressEnvironmentGetMsg struct {
	// Address to query
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty" validate:"required"`
}

type AddressEnvironmentGetResponseMsg struct {
	// Environment variables
	Environment map[string]string `json:"environment,omitempty" msgpack:"environment,omitempty" mapstructure:"environment,omitempty"`
}

type AddressEnvironmentSetMsg struct {
	// Address to query
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty" validate:"required"`

	// Environment variables
	Environment map[string]string `json:"environment" msgpack:"environment" mapstructure:"environment" validate:"required"`
}

type AddressEnvironmentSetResponseMsg struct {
	// Environment variables
	Environment map[string]string `json:"environment" msgpack:"environment" mapstructure:"environment"`
}

type AddressEnvironmentUnsetMsg struct {
	// Address to query
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty" validate:"required"`

	// Keys to remove
	Keys []string `json:"keys" msgpack:"keys" mapstructure:"keys" validate:"required"`
}

type AddressEnvironmentUnsetResponseMsg struct {
	// Environment variables
	Environment map[string]string `json:"environment" msgpack:"environment" mapstructure:"environment"`
}

/************/
/*  Events  */
/************/

type EventMsg struct {
	// SourceId is the unique identifier of the source of the event in the device
	SourceId string `bson:"sourceId,omitempty" json:"sourceId,omitempty" msgpack:"sourceId,omitempty" mapstructure:"sourceId,omitempty"`

	// UID must be provided by the client, and must be a unique identifier
	UID string `bson:"uid" json:"uid" msgpack:"uid" mapstructure:"uid,omitempty" validate:"required"`

	// Meta can hold any client provided data related to this event
	Meta map[string]interface{} `bson:"meta" json:"meta" msgpack:"meta" mapstructure:"meta,omitempty"`

	// Type parameter holds a mimetype or similar identifier of the payload
	Type string `bson:"type" json:"type" msgpack:"type" mapstructure:"type,omitempty"`

	// Payload is the raw data of the event
	Payload interface{} `bson:"payload" json:"payload" msgpack:"payload" mapstructure:"payload,omitempty" validate:"required"`

	// Domain is optional and can be used to create an event in a specific domain.
	// If left empty, the event will be created in the assigned domain of the address
	Domain string `bson:"domain" json:"domain" msgpack:"domain" mapstructure:"domain,omitempty"`
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

	// NoPayload will not return the payload of the events
	NoPayload bool `json:"noPayload" msgpack:"noPayload" mapstructure:"noPayload,omitempty"`

	// Type will filter out events that do not match the type
	Type string `json:"type" msgpack:"type" mapstructure:"type,omitempty"`
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
	Payload string `bson:"payload,omitempty" json:"payload,omitempty" msgpack:"payload,omitempty" mapstructure:"payload,omitempty" validate:"required"`
}

type SchemaResponseMsg struct {
	SchemaMsg `bson:",inline" mapstructure:",squash"`
	Hash      string `bson:"hash" json:"hash" msgpack:"hash" mapstructure:"hash,omitempty"`
}

type SchemaGetMsg struct {
	// Hash of the schema requested
	Hash string `bson:"hash" json:"hash" msgpack:"hash" mapstructure:"hash,omitempty" validate:"required"`

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
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty" validate:"required"`
}

type DomainDeleteMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty"  validate:"required"`
}

type DomainCreateMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty" validate:"required"`

	// Access rules (javascript snippet by default) to be applied to every message reaching an address in this domain
	AccessRules string `json:"accessRules" msgpack:"accessRules" mapstructure:"accessRules,omitempty"`

	// Variables added to the available environment during the rules execution
	Env map[string]string `json:"env" msgpack:"env" mapstructure:"env"`
}

type DomainCreateResponseMsg struct {
	// Domain name
	Domain `bson:",inline" mapstructure:",squash"`
}

type DomainUpdateMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty" validate:"required"`

	// Access rules (javascript snippet by default) to be applied to every message reaching an address in this domain
	AccessRules string `json:"accessRules" msgpack:"accessRules" mapstructure:"accessRules,omitempty"`

	// Variables added to the available environment during the rules execution
	Env map[string]string `json:"env" msgpack:"env" mapstructure:"env"`
}

type DomainUpdateResponseMsg struct {
	Domain `bson:",inline" mapstructure:",squash"`
}

type DomainUpdateAccessRulesMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty" validate:"required"`

	// Access rules (javascript snippet by default) to be applied to every message reaching an address in this domain
	AccessRules string `json:"accessRules" msgpack:"accessRules" mapstructure:"accessRules,omitempty" validate:"required"`
}

type DomainUpdateAccessRulesResponseMsg struct {
}

type DomainAssignMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty" validate:"required"`

	// Address to assign
	Address string `json:"address" msgpack:"address" mapstructure:"address,omitempty" validate:"required"`
}

type DomainGetTreeMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty" validate:"required"`
}

// TODO: transform to struct
type DomainGetTreeResponseMsg []string

type DomainCountAddressesMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty" validate:"required"`
}

type DomainCountAddressesResponseMsg struct {
	Addresses int `json:"addresses" msgpack:"addresses" mapstructure:"addresses,omitempty"`
}

type DomainListAddressesMsg struct {
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty" validate:"required"`
	Limit  uint   `json:"limit" msgpack:"limit" mapstructure:"limit,omitempty"`
	Skip   uint   `json:"skip" msgpack:"skip" mapstructure:"skip,omitempty"`
}

type DomainListAddressesResponseMsg struct {
	Addresses map[string]string `json:"addresses" msgpack:"addresses" mapstructure:"addresses,omitempty"`
}

type DomainEnvironmentGetMsg struct {
	// Domain to query
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty" validate:"required"`
}

type DomainEnvironmentGetResponseMsg struct {
	// Environment variables
	Environment map[string]string `json:"environment" msgpack:"environment" mapstructure:"environment"`
}

type DomainEnvironmentSetMsg struct {
	// Domain to query
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty" validate:"required"`

	// Environment variables
	Environment map[string]string `json:"environment,omitempty" msgpack:"environment,omitempty" mapstructure:"environment,omitempty"`
}

type DomainEnvironmentSetResponseMsg struct {
	// Environment variables
	Environment map[string]string `json:"environment" msgpack:"environment" mapstructure:"environment"`
}

type DomainEnvironmentUnsetMsg struct {
	// Domain to query
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty" validate:"required"`

	// Keys to remove
	Keys []string `json:"keys" msgpack:"keys" mapstructure:"keys" validate:"required"`
}

type DomainEnvironmentUnsetResponseMsg struct {
	// Environment variables
	Environment map[string]string `json:"environment" msgpack:"environment" mapstructure:"environment"`
}

/*************/
/*  Groups  */
/*************/

type GroupAddAddressMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain" validate:"required"`

	// Group name
	Group string `json:"group" msgpack:"group" mapstructure:"group" validate:"required"`

	// Address to assign
	Address string `json:"address" msgpack:"address" mapstructure:"address" validate:"required"`
}
type GroupAddAddressResponseMsg struct {
}

type GroupRemoveAddressMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain" validate:"required"`

	// Group name
	Group string `json:"group" msgpack:"group" mapstructure:"group" validate:"required"`

	// Address to remove
	Address string `json:"address" msgpack:"address" mapstructure:"address" validate:"required"`
}

type GroupRemoveAddressResponseMsg struct {
}

type GroupGetAddressesMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain" validate:"required"`

	// Group name
	Group string `json:"group" msgpack:"group" mapstructure:"group" validate:"required"`

	// True: Get addresses from all subdomains from Domain to tis TLD (bubble up)
	// False: Get addresses from all subdomains below Domain (propagate down)
	BubbleUp bool `json:"bubbleUp" msgpack:"bubbleUp" mapstructure:"bubbleUp"`
}

type GroupGetAddressesResponseMsg struct {
	// Addresses in group: domain -> addresses
	Addresses map[string][]string `json:"addresses" msgpack:"addresses" mapstructure:"addresses"`
}

type DomainGetGroupsMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain" validate:"required"`
}

type DomainGetGroupsResponseMsg struct {
	// Groups <domain>#<group>
	Groups []string `json:"groups" msgpack:"groups" mapstructure:"groups"`
}

type AddressGetGroupsMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty"`

	// Address
	Address string `json:"address" msgpack:"address" mapstructure:"address"  validate:"required"`
}

type AddressGetGroupsResponseMsg struct {
	// Address's groups: list of domain#group
	Groups []string `json:"groups" msgpack:"groups" mapstructure:"groups"`
}

type GroupRemoveMsg struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain" validate:"required"`

	// Group name
	Group string `json:"group" msgpack:"group" mapstructure:"group"  validate:"required"`

	// Propagate down the domain tree
	PropagateDown bool `json:"propagateDown" msgpack:"propagateDown" mapstructure:"propagateDown"`
}
type GroupRemoveResponseMsg struct {
}

type SessionCreateMsg struct {
	// Address to create session from
	Address string `json:"address" msgpack:"address" mapstructure:"address"`

	// Groups to assign to the session
	Groups []string `json:"groups" msgpack:"groups" mapstructure:"groups"`
}

type SessionCreateResponseMsg struct {
	// Created session ID
	Session string `json:"session" msgpack:"session" mapstructure:"session"`
}

type SessionDeleteMsg struct {
	// Session to delete
	Session string `json:"session" msgpack:"session" mapstructure:"session"`

	// Address to delete sessions from
	Address string `json:"address" msgpack:"address" mapstructure:"address"`
}

type SessionDeleteResponseMsg struct {
	// Sessions deleted
	Sessions int `json:"sessionsDeleted,omitempty" msgpack:"sessionsDeleted,omitempty" mapstructure:"sessionsDeleted"`
}

/*****************/
/*  Environment  */
/*****************/
type EnvironmentGetMsg struct {
	// Address to query environment of (empty for the current address)
	Address string `json:"address,omitempty" msgpack:"address,omitempty" mapstructure:"address,omitempty"`

	// Domain to query environment of (cannot be used with an address set)
	Domain string `json:"domain,omitempty" msgpack:"domain,omitempty" mapstructure:"domain,omitempty"`

	// Keys to request
	Keys []string `json:"keys,omitempty" msgpack:"keys,omitempty" mapstructure:"keys,omitempty"`
}

type EnvironmentGetResponseMsg struct {
	// Environment variables
	Environment map[string]string `json:"environment,omitempty" msgpack:"environment,omitempty" mapstructure:"environment,omitempty"`

	// Available keys
	AvailableKeys []string `json:"availableKeys,omitempty" msgpack:"availableKeys,omitempty" mapstructure:"availableKeys,omitempty"`
}
