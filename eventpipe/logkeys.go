package eventpipe

const (
	EventTimeField        = "time"
	DomainKey             = "domain"
	AddressKey            = "address"
	ErrorKey              = "error"
	EventKey              = "event"
	EventUIDKey           = EventKey + ".uid"
	EventPayloadSha256Key = EventKey + ".sha256"
	EventMetaKey          = EventKey + ".meta"
)
