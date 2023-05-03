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
	TopicCmdUpdateDevFirm = "efirm_updater.cmd.update"
)
