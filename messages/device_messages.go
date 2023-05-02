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
	Expr        string `bson:"expr" json:"expr" msgpack:"expr" mapstructure:"expr,omitempty"`
	FieldFilter string `bson:"fieldFilter" json:"fieldFilter" msgpack:"fieldFilter" mapstructure:"fieldFilter,omitempty"`
}

// list of usb devices and their usb attributes/env
type DevListResponseMsg struct {
	Devices []map[string]any `bson:"devices" json:"devices" msgpack:"devices" mapstructure:"devices"`
}
