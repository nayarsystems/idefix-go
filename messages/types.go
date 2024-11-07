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
	Address         string `mapstructure:"address" json:"address" msgpack:"address"`
	Product         string `mapstructure:"product" json:"product" msgpack:"product"`
	Board           string `mapstructure:"board" json:"board" msgpack:"board"`
	Version         string `mapstructure:"version" json:"version" msgpack:"version"`
	BootCnt         uint32 `mapstructure:"bootCnt" json:"bootCnt" msgpack:"bootCnt"`
	LauncherVersion string `mapstructure:"launcherVersion,omitempty" json:"launcherVersion,omitempty" msgpack:"launcherVersion,omitempty"`
}

type ConfigSyncInfo struct {
	Msg   string `json:"msg,omitempty" mapstructure:"msg,omitempty" msgpack:"msg,omitempty"`
	Error string `json:"error,omitempty" mapstructure:"error,omitempty" msgpack:"error,omitempty"`
}

type ConfigInfo struct {
	CloudFile       string         `json:"cloudFile,omitempty" mapstructure:"cloudFile,omitempty" msgpack:"cloudFile,omitempty"`
	CloudFileSha256 string         `json:"cloudFileSha256,omitempty" mapstructure:"cloudFileSha256,omitempty" msgpack:"cloudFileSha256,omitempty"`
	Dirty           bool           `json:"dirty" mapstructure:"dirty" msgpack:"dirty"`
	SyncInfo        ConfigSyncInfo `json:"syncInfo,omitempty" mapstructure:"syncInfo,omitempty" msgpack:"syncInfo,omitempty"`
}

type SysInfo struct {
	// Update SysInfoVersion in ToMsi method if you change this struct
	SysInfoVersion int `mapstructure:"sysInfoVersion,omitempty" json:"sysInfoVersion,omitempty" msgpack:"sysInfoVersion,omitempty"`

	DeviceInfo          `mapstructure:"devInfo" json:"devInfo" msgpack:"devInfo"`
	ConfigInfo          ConfigInfo    `mapstructure:"configInfo" json:"configInfo" msgpack:"configInfo"`
	LauncherErrorMsg    string        `mapstructure:"launchErr,omitempty" json:"launchErr,omitempty" msgpack:"launchErr,omitempty"`
	NumExecs            uint64        `mapstructure:"numExecs,omitempty" json:"numExecs,omitempty" msgpack:"numExecs,omitempty"`
	RollbackExec        bool          `mapstructure:"rollback,omitempty" json:"rollback,omitempty" msgpack:"rollback,omitempty"`
	SafeRunExec         bool          `mapstructure:"safeRun,omitempty" json:"safeRun,omitempty" msgpack:"safeRun,omitempty"`
	Uptime              time.Duration `mapstructure:"uptime,omitempty" json:"uptime,omitempty" msgpack:"uptime,omitempty"`
	LastRunUptime       time.Duration `mapstructure:"lastRunUptime,omitempty" json:"lastRunUptime,omitempty" msgpack:"lastRunUptime,omitempty"`
	LastRunExitCause    string        `mapstructure:"lastRunExitCause,omitempty" json:"lastRunExitCause,omitempty" msgpack:"lastRunExitCause,omitempty"`
	LastRunExitCode     int           `mapstructure:"lastRunExitCode,omitempty" json:"lastRunExitCode,omitempty" msgpack:"lastRunExitCode,omitempty"`
	LastRunExitIssuedBy string        `mapstructure:"lastRunExitIssuedBy,omitempty" json:"lastRunExitIssuedBy,omitempty" msgpack:"lastRunExitIssuedBy,omitempty"`
	LastRunExitIssuedAt time.Time     `mapstructure:"lastRunExitIssuedAt,omitempty" json:"lastRunExitIssuedAt,omitempty" msgpack:"lastRunExitIssuedAt,omitempty"`
}

func (m *SysInfo) ToMsi() (data msi, err error) {
	m.SysInfoVersion = 1
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
/*  States   */
/*************/

type StateEntry struct {
	Date       time.Time      `json:"date" msgpack:"date" mapstructure:"date" bson:"date"`
	BlobId     string         `json:"blobId" msgpack:"blobId" mapstructure:"blobId" bson:"blobId"`
	BlobMeta   map[string]any `json:"blobMeta" msgpack:"blobMeta" mapstructure:"blobMeta" bson:"blobMeta"`
	SchemaId   string         `json:"schemaId" msgpack:"schemaId" mapstructure:"schemaId" bson:"schemaId"`
	SchemaMeta map[string]any `json:"schemaMeta" msgpack:"schemaMeta" mapstructure:"schemaMeta" bson:"schemaMeta"`
	State      map[string]any `json:"state" msgpack:"state" mapstructure:"state" bson:"state"`
}

func (m *StateEntry) ToMsi() (data msi, err error) {
	data, err = ToMsiGeneric(m,
		mapstructure.ComposeEncodeFieldMapHookFunc(
			EncodeDurationToSecondsInt64Hook(),
			EncodeTimeToUnixMilliHook()))

	return data, err
}

func (m *StateEntry) ParseMsi(input msi) (err error) {
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
	Env map[string]interface{} `json:"env" msgpack:"env" mapstructure:"env"`

	// Creation time, hidden in Swagger
	Creation time.Time `json:"creation" msgpack:"creation" mapstructure:"-,omitempty" swagger:"-"`

	// Last update time, hidden in Swagger
	LastUpdate time.Time `json:"lastUpdate" msgpack:"lastUpdate" mapstructure:"-,omitempty" swagger:"-"`
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
