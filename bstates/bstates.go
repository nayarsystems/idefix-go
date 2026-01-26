package bstates

import (
	"encoding/base64"
	"fmt"
	"regexp"

	"github.com/nayarsystems/bstates"
	ifx "github.com/nayarsystems/idefix-go"
	"github.com/nayarsystems/idefix-go/messages"
)

func IsBstates(event *messages.Event) (yes bool, schemaId string) {
	schemaId, err := GetSchemaIdFromType(event.Type)
	return err == nil, schemaId
}

func GetSchemaFromEvent(ic *ifx.Client, event *messages.Event) (*bstates.StateSchema, error) {
	schemaId, err := GetSchemaIdFromType(event.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get bstates schema id from event type: %w", err)
	}

	schema, err := GetSchemaFromId(ic, schemaId)
	if err != nil {
		return nil, fmt.Errorf("failed to get bstates schema: %w", err)
	}

	return schema, nil
}

func GetStates(event *messages.Event, schema *bstates.StateSchema) ([]*bstates.State, error) {
	payload, err := normalizeEventPayload(event.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize event payload: %w", err)
	}

	stateQueue := bstates.CreateStateQueue(schema)
	if err := stateQueue.Decode(payload); err != nil {
		return nil, fmt.Errorf("failed to decode bstates payload: %w", err)
	}

	states, err := stateQueue.GetStates()
	if err != nil {
		return nil, fmt.Errorf("failed to get states from state queue: %w", err)
	}

	if len(states) == 0 {
		return nil, fmt.Errorf("no states in event")
	}

	return states, nil
}

func GetLastState(event *messages.Event, schema *bstates.StateSchema) (*bstates.State, error) {
	states, err := GetStates(event, schema)
	if err != nil {
		return nil, err
	}
	return states[len(states)-1], nil
}

// Gets the schema Id from a bstates based event's type field.
// Example:
//
// - input: "application/vnd.nayar.bstates; id=\"oEM5eJzBBGbyT9CLrSKrQwdnP2C+CVM8JHjfA0g3MAB=\""
//
// - output: "oEM5eJzBBGbyT9CLrSKrQwdnP2C+CVM8JHjfA0g3MAB="
func GetSchemaIdFromType(evtype string) (string, error) {
	r := regexp.MustCompile(`^application/vnd.nayar.bstates; id=([a-zA-Z0-9+/=]+)|"([a-zA-Z0-9+/=]+)"$`)

	matches := r.FindStringSubmatch(evtype)
	if matches == nil {
		return "", fmt.Errorf("no bstates type")
	}

	return matches[1] + matches[2], nil
}

func normalizeEventPayload(anyPayload any) ([]byte, error) {
	fromMsi := func(pv map[string]any) ([]byte, error) {
		payloadRaw, ok := pv["Data"]
		if !ok {
			return nil, fmt.Errorf("'Data' key not found in payload map")
		}
		return normalizeEventPayload(payloadRaw)
	}

	switch pv := anyPayload.(type) {
	case string:
		payload, err := base64.StdEncoding.DecodeString(pv)
		if err != nil {
			return nil, fmt.Errorf("payload is a string but is not valid base64: %v", err)
		}
		return payload, nil
	case []byte:
		// Already in the correct format
		return pv, nil
	case map[string]any:
		return fromMsi(pv)

	default:
		payloadMsi, err := messages.ToMsi(pv)
		if err != nil {
			return nil, fmt.Errorf("unexpected payload type: %T", pv)
		}
		return fromMsi(payloadMsi)
	}
}
