package eventpipe

import (
	"encoding/base64"
	"fmt"

	"github.com/nayarsystems/idefix-go/messages"
)

func normalizeEvent(event *messages.Event) error {
	fromMsi := func(pv map[string]any) error {
		payloadRaw, ok := pv["Data"]
		if !ok {
			return fmt.Errorf("'Data' key not found in payload map")
		}
		event.Payload = payloadRaw
		err := normalizeEvent(event)
		return err
	}

	switch pv := event.Payload.(type) {
	case string:
		payload, err := base64.StdEncoding.DecodeString(pv)
		if err != nil {
			return fmt.Errorf("'Data' is a string but is not valid base64: %v", err)
		}
		event.Payload = payload
	case []byte:
		// Already in the correct format
	case map[string]any:
		return fromMsi(pv)

	default:
		payloadMsi, err := messages.ToMsi(pv)
		if err != nil {
			return fmt.Errorf("unexpected payload type: %T", pv)
		}
		return fromMsi(payloadMsi)
	}

	return nil
}
