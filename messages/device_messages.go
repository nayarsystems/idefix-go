package messages

import "fmt"

/************/
/*   Info   */
/************/

type InfoReqMsg struct {
	Report          bool     `bson:"report" json:"report" msgpack:"report" mapstructure:"report"`
	ReportInstances []string `bson:"instances" json:"instances" msgpack:"instances" mapstructure:"instances,omitempty"`
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
	DEV_TYPE_RP2040 DevType = iota
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
