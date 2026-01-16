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

	// Serialize map[string]any to msgpack
	eventBytes, err := msgpack.Marshal(eventMap)
	if err != nil {
		slog.Error("failed to marshal event to msgpack", "error", err)
		return []byte{}
	}
	return eventBytes
}

func (e eventItem) ContextBytes() []byte {
	contextBytes, err := msgpack.Marshal(e.context)
	if err != nil {
		slog.Error("failed to marshal event context to msgpack", "error", err)
		return []byte{}
	}
	return contextBytes
}

func decodeContextBytes(contextBytes []byte) (map[string]any, error) {
	var context map[string]any
	if len(contextBytes) == 0 {
		return nil, nil
	}
	if err := msgpack.Unmarshal(contextBytes, &context); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context bytes: %w", err)
	}
	return context, nil
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
	return edb.St.DeleteItem(sourceId, eventId)
}

func (edb *EventsStorage) GetEvents(sourceId string) ([]*eventItem, error) {
	itemsList, err := edb.St.GetItems(sourceId, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get items from db: %w", err)
	}
	slog.Debug("retrieved items", "count", len(itemsList))
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

		context, err := decodeContextBytes(item.ContextBytes())
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
