package messages

import (
	"fmt"
	"time"

	"github.com/nayarsystems/mapstructure"
)

/************/
/*   Info   */
/************/

type DeviceInfo struct {
	Address         string `mapstructure:"address" json:"address"`
	Product         string `mapstructure:"product" json:"product"`
	Board           string `mapstructure:"board" json:"board"`
	Version         string `mapstructure:"version" json:"version"`
	BootCnt         uint32 `mapstructure:"bootCnt" json:"bootCnt"`
	LauncherVersion string `mapstructure:"launcherVersion,omitempty" json:"launcherVersion,omitempty"`
}

type SysInfo struct {
	DeviceInfo          `mapstructure:"devInfo"`
	LauncherErrorMsg    string        `mapstructure:"launchErr,omitempty" json:"launchErr,omitempty"`
	NumExecs            uint64        `mapstructure:"numExecs,omitempty" json:"numExecs,omitempty"`
	RollbackExec        bool          `mapstructure:"rollback,omitempty" json:"rollback,omitempty"`
	SafeRunExec         bool          `mapstructure:"safeRun,omitempty" json:"safeRun,omitempty"`
	Uptime              time.Duration `mapstructure:"uptime,omitempty" json:"uptime,omitempty"`
	LastRunUptime       time.Duration `mapstructure:"lastRunUptime,omitempty" json:"lastRunUptime,omitempty"`
	LastRunExitCause    string        `mapstructure:"lastRunExitCause,omitempty" json:"lastRunExitCause,omitempty"`
	LastRunExitCode     int           `mapstructure:"lastRunExitCode,omitempty" json:"lastRunExitCode,omitempty"`
	LastRunExitIssuedBy string        `mapstructure:"lastRunExitIssuedBy,omitempty" json:"lastRunExitIssuedBy,omitempty"`
	LastRunExitIssuedAt time.Time     `mapstructure:"lastRunExitIssuedAt,omitempty" json:"lastRunExitIssuedAt,omitempty"`
}

func (m *SysInfo) ToMsi() (data msi, err error) {
	data, err = ToMsiGeneric(m,
		mapstructure.ComposeEncodeFieldMapHookFunc(
			EncodeDurationToSecondsInt64Hook(),
			EncodeTimeToUnixMilliHook()))

	return data, err
}

func (m *SysInfo) ParseMsi(input msi) (err error) {
	err = ParseMsiGeneric(input, m,
		mapstructure.ComposeDecodeHookFunc(
			DecodeNumberToDurationHookFunc(time.Second),
			DecodeUnixMilliToTimeHookFunc()))
	if err != nil {
		return err
	}
	return nil
}

/*************/
/*  Domains  */
/*************/

type Domain struct {
	// Domain name
	Domain string `json:"domain" msgpack:"domain" mapstructure:"domain,omitempty" validate:"required"`

	// Access rules (javascript snippet by default) to be applied to every message reaching an address in this domain
	AccessRules string `json:"accessRules" msgpack:"accessRules" mapstructure:"accessRules,omitempty"`

	// Variables added to the available environment during the rules execution
	Env        map[string]interface{} `json:"env" msgpack:"env" mapstructure:"env"`
	Creation   time.Time              `json:"creation" msgpack:"creation" mapstructure:"-,omitempty"`
	LastUpdate time.Time              `json:"lastUpdate" msgpack:"lastUpdate" mapstructure:"-,omitempty"`
}

func (m *Domain) ToMsi() (data msi, err error) {
	data, err = ToMsiGeneric(m, EncodeTimeToStringHook(time.RFC3339))
	return
}

func (m *Domain) ParseMsi(input msi) (err error) {
	err = ParseMsiGeneric(input, m, DecodeAnyTimeStringToTimeHookFunc())
	return
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

func (m *Event) ToMsi() (data msi, err error) {
	data, err = ToMsiGeneric(m, EncodeTimeToStringHook(time.RFC3339))
	return
}

func (m *Event) ParseMsi(input msi) (err error) {
	err = ParseMsiGeneric(input, m, DecodeAnyTimeStringToTimeHookFunc())
	return
}

func (e *Event) String() string {
	return fmt.Sprintf("[%s] %s @ %s | %s: %v | %v", e.Timestamp, e.Address, e.Domain, e.Type, e.Payload, e.Meta)
}

/********************/
/*  Bstates events  */
/********************/

type SchemaInfo struct {
	Hash        string `bson:"_id" json:"hash" msgpack:"hash"`
	Address     string `bson:"address" json:"address" msgpack:"address"`
	Domain      string `bson:"domain" json:"domain" msgpack:"domain"`
	Description string `bson:"description" json:"description" msgpack:"description"`
	Payload     string `bson:"payload" json:"payload" msgpack:"payload"`
}

type Schema struct {
	SchemaInfo   `bson:",inline" mapstructure:",squash"`
	CreationTime time.Time `bson:"creationTime" json:"creationTime" msgpack:"creationTime"`
}

func (m *Schema) ToMsi() (data msi, err error) {
	data, err = ToMsiGeneric(m, EncodeTimeToStringHook(time.RFC3339))
	return
}

func (m *Schema) ParseMsi(input msi) (err error) {
	err = ParseMsiGeneric(input, m, DecodeAnyTimeStringToTimeHookFunc())
	return
}

/*********************/
/*   Binary Updates  */
/*********************/

type TargetExec = int

const (
	LauncherTargetExec TargetExec = iota
	IdefixTargetExec
)
