package messages

import (
	"fmt"
	"time"

	"github.com/jaracil/ei"
	"github.com/mitchellh/mapstructure"
)

/************/
/*   Info   */
/************/

type DeviceInfo struct {
	Address string `mapstructure:"address"`
	Product string `mapstructure:"product"`
	Board   string `mapstructure:"board"`
	Version string `mapstructure:"version"`
	BootCnt uint32 `mapstructure:"bootCnt"`
}

type SysInfo struct {
	DeviceInfo          `mapstructure:",squash"`
	LauncherErrorMsg    string        `mapstructure:"launcherr,omitempty"`
	LauncherFirstExec   bool          `mapstructure:"firstExec,omitempty"`
	RollbackExec        bool          `mapstructure:"rollback,omitempty"`
	Uptime              time.Duration `mapstructure:"uptime,omitempty"`
	LastRunUptime       time.Duration `mapstructure:"lastRunUptime,omitempty"`
	LastRunExitCause    string        `mapstructure:"lastRunExitCause,omitempty"`
	LastRunExitCode     int           `mapstructure:"lastRunExitCode,omitempty"`
	LastRunExitIssuedBy string        `mapstructure:"lastRunExitIssuedBy,omitempty"`
	LastRunExitIssuedAt time.Time     `mapstructure:"lastRunExitIssuedAt,omitempty"`
}

func (m *SysInfo) ToMsi() (data msi, err error) {
	data, err = ToMsiGeneric(m, nil)
	if err != nil {
		return nil, err
	}
	// time.Time fields has kind struct, so mapstructure encodes them as a map.
	// We need to convert them to unix milliseconds
	lastExitIssuedAt := m.LastRunExitIssuedAt.UnixMilli()
	if lastExitIssuedAt != 0 {
		data["lastRunExitIssuedAt"] = m.LastRunExitIssuedAt.UnixMilli()
	} else {
		delete(data, "lastRunExitIssuedAt")
	}

	// We want to encode uptime as seconds
	data["uptime"] = uint32(m.Uptime.Seconds())

	return data, err
}

func (m *SysInfo) ParseMsi(input msi) (err error) {
	err = ParseMsiGeneric(input, m,
		mapstructure.ComposeDecodeHookFunc(
			NumberToDurationHookFunc(time.Second),
			UnixMilliToTimeHookFunc()))
	if err != nil {
		return err
	}
	return nil
}

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
	data, err = ToMsiGeneric(m, nil)
	if err != nil {
		return nil, err
	}
	data["creation"] = TimeToString(m.Creation)
	data["lastUpdate"] = TimeToString(m.LastUpdate)
	return data, err
}

func (m *Domain) ParseMsi(input msi) (err error) {
	err = ParseMsiGeneric(input, m, nil)
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
	data, err = ToMsiGeneric(m, nil)
	if err != nil {
		return nil, err
	}
	// replace timeout field by its string format
	data["timestamp"] = TimeToString(m.Timestamp)
	return data, err
}

func (m *Event) ParseMsi(input msi) (err error) {
	m.Timestamp, _ = ei.N(input).M("timestamp").Time()
	err = ParseMsiGeneric(input, m, nil)
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
	data, err = ToMsiGeneric(m, nil)
	if err != nil {
		return nil, err
	}
	data["creationTime"] = TimeToString(m.CreationTime)
	return data, err
}

func (m *Schema) ParseMsi(input msi) (err error) {
	err = ParseMsiGeneric(input, m, nil)
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
