package messages

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
	DEVTYPE_RPI_PICO DevType = 0
)

type UpdateDevFirmReqMsg struct {
	DevType    DevType `bson:"devType" json:"devType" msgpack:"devType" mapstructure:"devType"`
	FirmUpdate []byte  `bson:"firmUpdate" json:"firmUpdate" msgpack:"firmUpdate" mapstructure:"firmUpdate"`
}

type UpdateDevFirmResMsg struct {
	Ok bool `json:"ok" msgpack:"ok" mapstructure:"ok"`
}
