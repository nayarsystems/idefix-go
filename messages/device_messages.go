package messages

import (
	"fmt"
	"time"

	"github.com/nayarsystems/mapstructure"
)

/******************/
/*   Idefix Exit  */
/******************/

type ExitReqMsg struct {
	Source        string        `mapstructure:"source"`
	StopDelay     time.Duration `mapstructure:"stopDelay"`
	WaitHaltDelay time.Duration `mapstructure:"waitHaltDelay"`
	ExitCode      int           `mapstructure:"exitCode"`
	ExitCause     string        `mapstructure:"exitCause"`
}

func (m *ExitReqMsg) ToMsi() (data msi, err error) {
	data, err = ToMsiGeneric(m, EncodeDurationToSecondsInt64Hook())
	return data, err
}

func (m *ExitReqMsg) ParseMsi(input msi) (err error) {
	err = ParseMsiGeneric(input, m, DecodeNumberToDurationHookFunc(time.Second))
	if err != nil {
		return err
	}
	return nil
}

/**************/
/*   SysInfo  */
/**************/

type SysInfoReqMsg struct {
	Report          bool     `bson:"report" json:"report" msgpack:"report" mapstructure:"report"`
	ReportInstances []string `bson:"instances" json:"instances" msgpack:"instances" mapstructure:"instances,omitempty"`
}

type SysInfoResMsg struct {
	SysInfo `mapstructure:",squash"`
	Report  map[string]map[string]interface{} `mapstructure:"report,omitempty"`
}

func (m *SysInfoResMsg) ToMsi() (data msi, err error) {
	data, err = ToMsiGeneric(m,
		mapstructure.ComposeEncodeFieldMapHookFunc(
			EncodeDurationToSecondsInt64Hook(),
			EncodeTimeToUnixMilliHook()))

	return data, err
}

func (m *SysInfoResMsg) ParseMsi(input msi) (err error) {
	err = ParseMsiGeneric(input, m,
		mapstructure.ComposeDecodeHookFunc(
			DecodeNumberToDurationHookFunc(time.Second),
			DecodeUnixMilliToTimeHookFunc()))
	if err != nil {
		return err
	}
	return nil
}

/************/
/*   Exec   */
/************/

type ExecReqMsg struct {
	Cmd string `bson:"command" json:"command" msgpack:"command" mapstructure:"command"`
}

type ExecResMsg struct {
	Code    int    `bson:"code" json:"code" msgpack:"code" mapstructure:"code"`
	Stdout  string `bson:"stdout" json:"stdout" msgpack:"stdout" mapstructure:"stdout"`
	Stderr  string `bson:"stderr" json:"stderr" msgpack:"stderr" mapstructure:"stderr"`
	Success bool   `bson:"success" json:"success" msgpack:"success" mapstructure:"success"`
}

/***************/
/*   OS Utils  */
/***************/

type FileWriteMsg struct {
	// Path to write the file
	Path string `mapstructure:"path,omitempty"`

	// File content
	Data []byte `mapstructure:"data,omitempty"`

	// File permissions
	Mode uint32 `mapstructure:"mode,omitempty"`
}

type FileWriteResMsg struct {
	// File hash
	Hash []byte `mapstructure:"hash,omitempty"`
}

type FileReadMsg struct {
	// Path to write the file
	Path string `mapstructure:"path,omitempty"`
}

type FileReadResMsg struct {
	// File content
	Data []byte `mapstructure:"data,omitempty"`
}

type FileSHA256Msg struct {
	// Path to write the file
	Path string `mapstructure:"path,omitempty"`
}

type FileSHA256ResMsg struct {
	// File hash
	Hash []byte `mapstructure:"hash,omitempty"`
}

type FileSizeMsg struct {
	// File path to check size
	Path string `mapstructure:"path,omitempty"`
}

type FileSizeResMsg struct {
	// File size in bytes
	Size int64 `mapstructure:"size,omitempty"`
}

type FreeSpaceMsg struct {
	// Directory path to check free space. If empty, check idefix's working directory
	Path string `mapstructure:"path,omitempty"`
}

type FreeSpaceResMsg struct {
	// Bytes of free space
	Free uint64 `mapstructure:"free,omitempty"`
}

type RemoveMsg struct {
	// File path to remove
	Path string `mapstructure:"path,omitempty"`
}

type RemoveResMsg struct {
}

type FileCopyMsg struct {
	// Source path
	SrcPath string `mapstructure:"srcPath,omitempty"`

	// Dst path
	DstPath string `mapstructure:"dstPath,omitempty"`
}

type FileCopyResMsg struct {
}

type MoveMsg struct {
	// Source path
	SrcPath string `mapstructure:"srcPath,omitempty"`

	// Dst path
	DstPath string `mapstructure:"dstPath,omitempty"`
}

type MoveResMsg struct {
}

const (
	MkdirTypeVolatile = 0
	MkdirTypeScratch  = 1
	MkdirTypeAbsolute = 2
)

type MkdirMsg struct {
	Type int    `mapstructure:"type,omitempty"`
	Path string `mapstructure:"path,omitempty"`
}

type MkdirResMsg struct {
	Path string `mapstructure:"path,omitempty"`
}

type PatchMsg struct {
	Source      string `mapstructure:"source,omitempty"`
	Destination string `mapstructure:"destination,omitempty"`
	PatchData   []byte `mapstructure:"patchData,omitempty"`
	PatchPath   string `mapstructure:"patchPath,omitempty"`
}

type PatchResMsg struct {
}

type ListDirMsg struct {
	Path string `mapstructure:"path,omitempty"`
}

type FileInfo struct {
	Name  string `mapstructure:"name,omitempty"`
	IsDir bool   `mapstructure:"isDir,omitempty"`
	Size  int64  `mapstructure:"size,omitempty"`
	Mode  uint32 `mapstructure:"mode,omitempty"`
}

type ListDirResMsg struct {
	Files []*FileInfo
}

/******************/
/*   Idefix update  */
/******************/

const (
	UpdateTypeIdefixUpgrade   = 0
	UpdateTypeIdefixRollback  = 1
	UpdateTypeLauncherUpgrade = 2
)

type UpdateMsg struct {
	Type          int           `mapstructure:"type"`
	Cause         string        `mapstructure:"cause,omitempty"`
	StopDelay     time.Duration `mapstructure:"stopDelay,omitempty"`
	WaitHaltDelay time.Duration `mapstructure:"waitHaltDelay,omitempty"`
}

func (m *UpdateMsg) ToMsi() (data msi, err error) {
	data, err = ToMsiGeneric(m, EncodeDurationToSecondsInt64Hook())
	return data, err
}

func (m *UpdateMsg) ParseMsi(input msi) (err error) {
	err = ParseMsiGeneric(input, m, DecodeNumberToDurationHookFunc(time.Second))
	if err != nil {
		return err
	}
	return nil
}

type UpdateResMsg struct {
	UpdateMsg `mapstructure:",squash"`
}

/*******************/
/*   Udev module   */
/*******************/

type DevListReqMsg struct {
	Expr        string `bson:"expr" json:"expr" msgpack:"expr" mapstructure:"expr"`
	FieldFilter string `bson:"fieldFilter" json:"fieldFilter" msgpack:"fieldFilter" mapstructure:"fieldFilter,omitempty"`
}

// list of usb devices and their usb attributes/env
type DevListResponseMsg struct {
	Devices []map[string]any `bson:"devices" json:"devices" msgpack:"devices" mapstructure:"devices"`
}

/********************************************************/
/*   Manage firmware updates over devices physically    */
/*   connected to a idefix client (e.g: RP2040, STM32)  */
/********************************************************/

type DevType int

const (
	DEV_TYPE_UNKNOWN DevType = iota
	DEV_TYPE_RP2040
)

type UpdateFileType int

const (
	UPDATE_FILE_TYPE_UNSPECIFIED UpdateFileType = iota
	UPDATE_FILE_TYPE_BIN
	UPDATE_FILE_TYPE_UF2
	UPDATE_FILE_TYPE_ELF
	UPDATE_FILE_TYPE_HEX
	UPDATE_FILE_TYPE_TAR
)

func ParseDevType(input string) (fileType DevType, err error) {
	switch input {
	case "rp2040":
		fileType = DEV_TYPE_RP2040
	default:
		err = fmt.Errorf("unkown type '%s'", input)
	}
	return
}

func ParseFileType(input string) (fileType UpdateFileType, err error) {
	switch input {
	case "bin":
		fileType = UPDATE_FILE_TYPE_BIN
	case "uf2":
		fileType = UPDATE_FILE_TYPE_UF2
	case "elf":
		fileType = UPDATE_FILE_TYPE_ELF
	case "hex":
		fileType = UPDATE_FILE_TYPE_HEX
	case "tar":
		fileType = UPDATE_FILE_TYPE_TAR
	default:
		err = fmt.Errorf("unkown type '%s'", input)
	}
	return
}

// UsbPort and UsbPath are mutually exclusive and not all devices will require them (e.g: SPI or I2C connected devices)
// Some devices devices may require a custom file type. For these cases FileType will be optional and can be set to UPDATE_FILE_TYPE_UNSPECIFIED.
type UpdateDevFirmReqMsg struct {
	DevType  DevType        `bson:"devType" json:"devType" msgpack:"devType" mapstructure:"devType"`
	UsbPort  string         `bson:"usbPort" json:"usbPort" msgpack:"usbPort" mapstructure:"usbPort,omitempty"`
	UsbPath  string         `bson:"usbPath" json:"usbPath" msgpack:"usbPath" mapstructure:"usbPath,omitempty"`
	FileType UpdateFileType `bson:"fileType" json:"fileType" msgpack:"fileType" mapstructure:"fileType"`
	File     []byte         `bson:"file" json:"file" msgpack:"file" mapstructure:"file"`
	FileHash []byte         `bson:"hash" json:"hash" msgpack:"hash" mapstructure:"hash"`
}

// Contains the update command's output
type UpdateDevFirmResMsg struct {
	Output string `json:"output" msgpack:"output" mapstructure:"output"`
}

// UsbPort and UsbPath are mutually exclusive and not all devices will require them (e.g: SPI or I2C connected devices)
type RebootDevReqMsg struct {
	DevType DevType `bson:"devType" json:"devType" msgpack:"devType" mapstructure:"devType"`
	UsbPort string  `bson:"usbPort" json:"usbPort" msgpack:"usbPort" mapstructure:"usbPort,omitempty"`
	UsbPath string  `bson:"usbPath" json:"usbPath" msgpack:"usbPath" mapstructure:"usbPath,omitempty"`
}

// Contains the reboot device command's output
type RebootDevResMsg struct {
	Output string `json:"output" msgpack:"output" mapstructure:"output"`
}
