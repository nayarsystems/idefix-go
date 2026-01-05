package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nayarsystems/bstates"
	"github.com/nayarsystems/idefix-go/messages"
)

var schemasCache = map[string]*bstates.StateSchema{}

// Process the event here
func processBstatesEvent(event *messages.Event) error {
	slog.Info("parsing bstates event", "id", event.UID, "domain", event.Domain, "address", event.Address, "timestamp", event.Timestamp,
		"sourceId", event.SourceId, "meta", event.Meta)

	schemaId, err := messages.BstatesParseSchemaIdFromType(event.Type)
	if err != nil {
		return fmt.Errorf("failed to parse bstates schema id: %w", err)
	}

	slog.Info("bstates schema id", "schemaId", schemaId)
	schema, err := getSchema(schemaId)
	if err != nil {
		return fmt.Errorf("failed to get bstates schema: %w", err)
	}

	payload, err := normalizePayload(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to normalize payload: %w", err)
	}

	// Decode bstates blob using the schema
	// First create an empty state queue that follows the schema
	stateQueue := bstates.CreateStateQueue(schema)

	// Fill the state queue with the payload
	if err := stateQueue.Decode(payload); err != nil {
		return fmt.Errorf("failed to decode bstates payload: %w", err)
	}

	// Blob parsed successfully.
	states, err := stateQueue.GetStates()
	if err != nil {
		return fmt.Errorf("failed to get states from state queue: %w", err)
	}

	slog.Info("bstates event parsed successfully", "numStates", len(states))

	for i, state := range states {
		// Convert state to map[string]any (msi) to easily convert to json
		stateMsi, err := state.ToMsi()
		if err != nil {
			return fmt.Errorf("failed to convert state to msi: %w", err)
		}

		stateJson, err := json.MarshalIndent(stateMsi, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal state msi to json: %w", err)
		}

		fmt.Printf("State %d:\n%s\n", i, string(stateJson))
	}
	return nil
}

func getSchema(schemaId string) (*bstates.StateSchema, error) {
	// Check cache first
	if schema, ok := schemasCache[schemaId]; ok {
		return schema, nil
	}

	// Not in cache, fetch it
	slog.Info("fetching bstates schema", "schemaId", schemaId)
	res, err := client.GetSchema(schemaId, time.Second*20)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bstates schema: %w", err)
	}

	// Create schema object
	schema := &bstates.StateSchema{}
	err = schema.UnmarshalJSON([]byte(res.Payload))
	if err != nil {
		return nil, fmt.Errorf("failed to parse bstates schema: %w", err)
	}

	// Cache it for future use
	schemasCache[schemaId] = schema
	return schema, nil
}

func normalizePayload(anyPayload any) ([]byte, error) {
	fromMsi := func(pv map[string]any) ([]byte, error) {
		payloadRaw, ok := pv["Data"]
		if !ok {
			return nil, fmt.Errorf("'Data' key not found in payload map")
		}
		return normalizePayload(payloadRaw)
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
