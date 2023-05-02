package messages

const (
	// TopicCmdGetDevInfo is used to get the entire ENV for every device matched with the given mongo expression
	//
	// - message: DevListReqMsg
	//
	// - response: DevListResponseMsg
	TopicCmdGetDevInfo = "udev.cmd.info"
)
