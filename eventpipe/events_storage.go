package eventpipe

import (
	"fmt"
	"log/slog"

	"github.com/nayarsystems/idefix-go/eventpipe/storage"
	"github.com/nayarsystems/idefix-go/messages"
	"github.com/vmihailenco/msgpack/v5"
)

type EventsStorage struct {
	St storage.Storage
}

type eventItem struct {
	sourceId string
	*messages.Event
	context map[string]any
}

func (e eventItem) SourceId() string {
	return e.sourceId
}

func (e eventItem) Id() string {
	return e.Event.UID
}

func (e eventItem) Bytes() []byte {
	// Convert Event to map[string]any using ToMsi
	eventMap, err := e.Event.ToMsi()
	if err != nil {
		slog.Error("failed to convert event to map", "error", err)
		return []byte{}
	}

	res, err := encodeMsi(eventMap)
	if err != nil {
		slog.Error("failed to serialize event", "error", err)
		return []byte{}
	}
	return res
}

func (e eventItem) ContextBytes() []byte {
	contextBytes, err := encodeMsi(e.context)
	if err != nil {
		slog.Error("failed to serialize event context", "error", err)
		return []byte{}
	}
	return contextBytes
}

func encodeMsi(input map[string]any) ([]byte, error) {
	if input == nil {
		return []byte{}, nil
	}
	return msgpack.Marshal(input)
}

func decodeMsi(contextBytes []byte) (map[string]any, error) {
	var context map[string]any
	if len(contextBytes) == 0 {
		return nil, nil
	}
	if err := msgpack.Unmarshal(contextBytes, &context); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context bytes: %w", err)
	}
	return context, nil
}

func normalizeMap(input map[string]any) (map[string]any, error) {
	if input == nil {
		return map[string]any{}, nil
	}
	bytes, err := encodeMsi(input)
	if err != nil {
		return nil, fmt.Errorf("failed to encode map: %w", err)
	}
	output, err := decodeMsi(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode map: %w", err)
	}
	return output, nil
}

func (edb *EventsStorage) PushEvent(sourceId string, event *messages.Event, context map[string]any) error {
	// Create an eventItem
	item := eventItem{
		sourceId: sourceId,
		Event:    event,
		context:  context,
	}
	// Push the event to the database (index auto-assigned)
	return edb.St.Push(item)
}

func (edb *EventsStorage) UpdateEvent(sourceId string, event *messages.Event, context map[string]any) error {
	// Create an eventItem
	item := eventItem{
		sourceId: sourceId,
		Event:    event,
		context:  context,
	}
	// Update the event in the database
	return edb.St.Update(item)
}

func (edb *EventsStorage) DeleteEvent(sourceId, eventId string) error {
	return edb.St.Delete(sourceId, eventId)
}

func (edb *EventsStorage) GetEvents(sourceId string) ([]*eventItem, error) {
	itemsList, err := edb.St.GetItems(sourceId, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get items from db: %w", err)
	}
	events := []*eventItem{}
	for _, item := range itemsList {
		// Deserialize msgpack to map[string]any
		var eventMap map[string]any
		if err := msgpack.Unmarshal(item.Bytes(), &eventMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msgpack: %w", err)
		}

		// Convert map[string]any back to Event using ParseMsi
		var event messages.Event
		if err := event.ParseMsi(eventMap); err != nil {
			return nil, fmt.Errorf("failed to parse event from map: %w", err)
		}

		context, err := decodeMsi(item.ContextBytes())
		if err != nil {
			return nil, fmt.Errorf("failed to decode context bytes: %w", err)
		}

		eItem := eventItem{
			sourceId: item.SourceId(),
			Event:    &event,
			context:  context,
		}
		events = append(events, &eItem)
	}
	return events, nil
}

func (edb *EventsStorage) GetUnlockedEvents(sourceId string, limit int) ([]*eventItem, error) {
	itemsList, err := edb.St.GetUnlocked(sourceId, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get unlocked items from db: %w", err)
	}
	events := []*eventItem{}
	for _, item := range itemsList {
		// Deserialize msgpack to map[string]any
		var eventMap map[string]any
		if err := msgpack.Unmarshal(item.Bytes(), &eventMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msgpack: %w", err)
		}

		// Convert map[string]any back to Event using ParseMsi
		var event messages.Event
		if err := event.ParseMsi(eventMap); err != nil {
			return nil, fmt.Errorf("failed to parse event from map: %w", err)
		}

		context, err := decodeMsi(item.ContextBytes())
		if err != nil {
			return nil, fmt.Errorf("failed to decode context bytes: %w", err)
		}

		eItem := eventItem{
			sourceId: item.SourceId(),
			Event:    &event,
			context:  context,
		}
		events = append(events, &eItem)
	}
	return events, nil
}

func (edb *EventsStorage) LockEvent(sourceId, eventId string) error {
	return edb.St.Lock(sourceId, eventId)
}

func (edb *EventsStorage) UnlockEvent(sourceId, eventId string) error {
	return edb.St.Unlock(sourceId, eventId)
}

func (edb *EventsStorage) UnlockAllEvents(sourceId string) error {
	lockedItems, err := edb.St.GetLocked(sourceId, 0)
	if err != nil {
		return fmt.Errorf("failed to get locked items from db: %w", err)
	}
	for len(lockedItems) > 0 {
		for _, item := range lockedItems {
			err := edb.St.Unlock(sourceId, item.Id())
			if err != nil {
				return fmt.Errorf("failed to unlock item with id '%s': %w", item.Id(), err)
			}
		}
		lockedItems, err = edb.St.GetLocked(sourceId, 0)
		if err != nil {
			return fmt.Errorf("failed to get locked items from db: %w", err)
		}

	}
	return nil
}
