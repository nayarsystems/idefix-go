package messages

const (
	// TopicRemoteSubscribe is the prefix for remote subscription using public MQTT topic
	//
	// - message: StreamSubMsg
	//
	// - response: StreamSubResMsg
	TopicRemoteSubscribe = "streams.cmd.sub"

	TopicRemoteStartPublisher = "streams.cmd.start_pub"

	// TopicRemoteUnsubscribe is the prefix for remote unsubscription using public MQTT topic
	//
	// - message: StreamUnsubMsg
	//
	// - response: StreamUnsubResMsg
	TopicRemoteUnsubscribe = "streams.cmd.unsub"

	TopicRemoteStopPublisher = "streams.cmd.stop_pub"

	// TopicCmdGetDevInfo is used to get the entire ENV for every device matched with the given mongo expression
	//
	// - message: DevListReqMsg
	//
	// - response: DevListResponseMsg
	TopicCmdGetDevInfo = "udev.cmd.info"

	// TopicCmdExec is used to execute a command on device
	//
	// - message: ExecReqMsg
	//
	// - response: ExecResMsg
	TopicCmdExec = "os.cmd.exec"

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
