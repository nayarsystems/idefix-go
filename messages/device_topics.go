package messages

const (
	// TopicCmdGetDevInfo is used to get the entire ENV for every device matched with the given mongo expression
	//
	// - message: DevListReqMsg
	//
	// - response: DevListResponseMsg
	TopicCmdGetDevInfo = "udev.cmd.info"

	// TopicCmdUpdateDevFirm is used to update the firmware
	// of a device physically connected to a idefix client (e.g: RP2040, STM32)
	//
	// - message: UpdateDevFirmReqMsg
	//
	// - response: UpdateDevFirmResMsg
	TopicCmdUpdateDevFirm = "edev_manager.cmd.update"

	// TopicCmdRebootDev2Flash is used to reboot a device to flash mode
	// physically connected to a idefix client (e.g: RP2040, STM32)
	//
	// - message: RebootDevReqMsg
	//
	// - response: RebootDevResMsg
	TopicCmdRebootDev2Flash = "edev_manager.cmd.reboot2flash"

	// TopicCmdRebootDev2App is used to reboot a device to app mode
	// physically connected to a idefix client (e.g: RP2040, STM32)
	//
	// - message: RebootDevReqMsg
	//
	// - response: RebootDevResMsg
	TopicCmdRebootDev2App = "edev_manager.cmd.reboot2app"

	// TopicUsbEvtPathPrefix is used to subscribe to up/down (attached/detached) events on a usb path
	//
	// - format: usb.evt.path.<path>
	TopicUsbEvtPathPrefix = "usb.evt.path"

	// TopicUsbEvtPortPrefix is used to subscribe to up/down (attached/detached) events on a usb port
	//
	// - format: usb.evt.port.<port>
	TopicUsbEvtPortPrefix = "usb.evt.port"
)

func TopicUsbEvtPath(path string) string {
	return TopicUsbEvtPathPrefix + "." + path
}

func TopicUsbEvtPort(port string) string {
	return TopicUsbEvtPortPrefix + "." + port
}