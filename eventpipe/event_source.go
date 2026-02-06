package eventpipe

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/go-pipeline/pkg/pipeline"
	"github.com/jaracil/ei"
	ierrors "github.com/nayarsystems/idefix-go/errors"
	m "github.com/nayarsystems/idefix-go/messages"
)

type EventSource struct {
	l          *slog.Logger
	m          *EventSourceManager
	p          EventSourceParams
	stages     []pipeline.Stage[pipelineItem]
	stageNames map[string]struct{}
}

type EventSourceParams struct {
	// Query events since this time if no cursor found in storage
	// If zero, query all events (limited by service query limits)
	Since time.Time

	// If specified, continue fetching events from this cursor
	// if no cursor found in storage
	ContinuationID string

	// Domain of the source
	Domain string

	// (optional) Address filter
	Address string

	// (optional) Type filter
	Type string

	// (optional) Custom id for the source.
	// If empty, id will be generated based on other parameters
	Id string

	// Long polling timeout for event fetching
	LongPollingTimeout time.Duration
}

type pipelineItem struct {
	event        *m.Event
	eventContext map[string]any
	passthrough  bool
}

type eventStageOptions struct {
	name        string // name of the stage
	concurrency uint   // number of concurrent workers
	bufferSize  uint   // size of the input buffer
}

func (s *EventSource) Push(stage EventStage, options ...EventStageOptionFn) (err error) {
	var eventStageOptions eventStageOptions
	eventStageOptions.concurrency = 1
	eventStageOptions.bufferSize = 1

	for _, optFn := range options {
		if err := optFn(&eventStageOptions); err != nil {
			return err
		}
	}

	gopts := []pipeline.StageOptionFn{}
	if eventStageOptions.name == "" {
		eventStageOptions.name = fmt.Sprintf("__s%d", len(s.stages))
	}
	if s.stageNames == nil {
		s.stageNames = make(map[string]struct{})
	}
	if _, exists := s.stageNames[eventStageOptions.name]; exists {
		return fmt.Errorf("stage with name '%s' already exists in the pipeline", eventStageOptions.name)
	}
	gopts = append(gopts, pipeline.Name(eventStageOptions.name))
	gopts = append(gopts, pipeline.Concurrency(eventStageOptions.concurrency))
	gopts = append(gopts, pipeline.InputBufferSize(eventStageOptions.bufferSize))

	stageIndex := len(s.stages)

	gstage := pipeline.NewStage(
		func(in pipelineItem) (out pipelineItem, err error) {
			if s.m.ctx.Err() != nil {
				return in, s.m.ctx.Err()
			}
			isLastStage := stageIndex == len(s.stages)-1
			if in.passthrough {
				s.l.Debug("passing through event", "event_id", in.event.UID, "stage", eventStageOptions.name)
				return in, nil
			}
			eventContext := in.eventContext
			if eventContext == nil {
				eventContext = make(map[string]any)
			}
			processedStages := ei.N(eventContext).M("processedStages").MapStrZ()
			if processedStages == nil {
				processedStages = make(map[string]any)
				eventContext["processedStages"] = processedStages
			}
			if ei.N(processedStages).M(eventStageOptions.name).BoolZ() {
				s.l.Debug("skipping event already processed in stage", "event_id", in.event.UID, "stage", eventStageOptions.name)
				out = in
				out.eventContext = eventContext
				return out, nil
			}
			pipelineContext := ei.N(eventContext).M("pipelineContext").MapStrZ()
			if pipelineContext == nil {
				pipelineContext = make(map[string]any)
			}
			stageInput := EventStageInput{
				Event:           in.event,
				PipelineContext: pipelineContext,
			}
			var stageOutput EventStageOutput
			s.l.Debug("processing event in stage", "event_id", in.event.UID, "stage", eventStageOptions.name, "pipeline_context", stageInput.PipelineContext)
			stageOutput, err = stage.Process(s.m.ctx, stageInput)
			if err != nil {
				return out, err
			}
			out.event = stageOutput.Event
			if out.event == nil {
				out.event = in.event
			}

			s.l.Debug("stage processing completed", "event_id", in.event.UID, "stage", eventStageOptions.name, "pipeline_context", stageOutput.PipelineContext)
			if stageOutput.Remove || (stageOutput.Processed && isLastStage) {
				s.deleteEvent(in.event)
				out.passthrough = true
				s.l.Debug("event removed from pipeline", "event_id", in.event.UID, "stage", eventStageOptions.name)
				return out, nil
			}
			if stageOutput.PipelineContext == nil {
				stageOutput.PipelineContext = make(map[string]any)
			}

			eventContext["pipelineContext"] = stageOutput.PipelineContext
			processedStages[eventStageOptions.name] = stageOutput.Processed

			out.eventContext, err = normalizeMap(eventContext)
			if err != nil {
				return out, err
			}
			if err = s.updateEvent(out.event, out.eventContext); err != nil {
				return out, err
			}
			if isLastStage || !stageOutput.Processed {
				// Unlock the event for future processing
				if err = s.unlockEvent(out.event); err != nil {
					return out, err
				}
				out.passthrough = true
				s.l.Debug("event unlocked for future processing", "event_id", out.event.UID, "stage", eventStageOptions.name)
			} else {
				s.l.Debug("event processed in stage", "event_id", out.event.UID, "stage", eventStageOptions.name)
			}
			return out, nil
		},
		gopts...,
	)
	s.stages = append(s.stages, gstage)
	s.stageNames[eventStageOptions.name] = struct{}{}
	return nil
}

func (s *EventSource) buildProducer() pipeline.Producer[pipelineItem] {
	producerFn := func(put func(pipelineItem)) error {
		defer s.m.cancel(nil)
		err := s.producerFunc(put)
		if err != nil {
			s.m.cancel(err)
			return err
		}
		return nil
	}
	producer := pipeline.NewProducer(producerFn, pipeline.Name("event_source_producer"))
	return producer
}

func (s *EventSource) Run() (err error) {
	err = pipeline.Do(s.buildProducer(), s.stages...)
	s.m.cancel(err)
	return err
}

func (s *EventSource) RunAndMeasure() (stats *pipeline.Metrics, err error) {
	stats, err = pipeline.Measure(s.buildProducer(), s.stages...)
	s.m.cancel(err)
	return stats, err
}

func (s *EventSource) producerFunc(put func(pipelineItem)) error {
	err := s.unlockAllEvents()
	if err != nil {
		return fmt.Errorf("failed to unlock events: %w", err)
	}

	var events []*m.Event
	cursor, err := s.m.st.GetCursor(s.Id())
	if err != nil {
		cursor = s.p.ContinuationID
	}

	for s.waitWithContext(time.Second) == nil {
		storageEvents, err := s.loadUnlockedEvents()
		if err != nil {
			return fmt.Errorf("failed to load pending events: %w", err)
		}
		s.l.Info("loaded pending events from storage", "count", len(storageEvents), "cursor", cursor)
		for _, e := range storageEvents {
			if err := s.lockEvent(e.Event); err != nil {
				return fmt.Errorf("failed to lock event: %w", err)
			}
			put(pipelineItem{
				event:        e.Event,
				eventContext: e.context,
			})
		}

		if s.m.ctx.Err() != nil {
			return s.m.ctx.Err()
		}

		events, cursor, err = s.fetchEvents(cursor)
		if err != nil {
			return fmt.Errorf("failed to fetch events: %w", err)
		}
		s.l.Info("fetched events from source", "count", len(events), "cursor", cursor)

		for _, e := range events {
			if err := s.pushEvent(e, nil); err != nil {
				return fmt.Errorf("failed to push event to db: %w", err)
			}
		}

		if len(events) > 0 {
			err = s.m.st.UpdateCursor(s.Id(), cursor)
			if err != nil {
				return fmt.Errorf("failed to update cursor: %w", err)
			}
		}
	}
	return nil
}

func (s *EventSource) pushEvent(e *m.Event, meta map[string]any) (err error) {
	db := EventsStorage{St: s.m.st}
	return db.PushEvent(s.Id(), e, meta)
}

func (s *EventSource) updateEvent(e *m.Event, meta map[string]any) (err error) {
	db := EventsStorage{St: s.m.st}
	return db.UpdateEvent(s.Id(), e, meta)
}

func (s *EventSource) deleteEvent(e *m.Event) (err error) {
	db := EventsStorage{St: s.m.st}
	return db.DeleteEvent(s.Id(), e.UID)
}

func (s *EventSource) loadUnlockedEvents() ([]*eventItem, error) {
	db := EventsStorage{St: s.m.st}
	items, err := db.GetUnlockedEvents(s.Id(), 100)
	return items, err
}

func (s *EventSource) unlockAllEvents() error {
	db := EventsStorage{St: s.m.st}
	err := db.UnlockAllEvents(s.Id())
	return err
}

func (s *EventSource) lockEvent(e *m.Event) error {
	db := EventsStorage{St: s.m.st}
	err := db.LockEvent(s.Id(), e.UID)
	return err
}

func (s *EventSource) unlockEvent(e *m.Event) error {
	db := EventsStorage{St: s.m.st}
	err := db.UnlockEvent(s.Id(), e.UID)
	return err
}

func (s *EventSource) fetchEvents(cursor string) ([]*m.Event, string, error) {
	queryContext, queryCancel := context.WithTimeout(s.m.ctx, s.p.LongPollingTimeout+10*time.Second)
	defer queryCancel()
	res, err := s.m.c.EventsGet(&m.EventsGetMsg{
		Domain:         s.p.Domain,
		Address:        s.p.Address,
		Since:          s.p.Since,
		Timeout:        s.p.LongPollingTimeout,
		ContinuationID: cursor,
		Limit:          100,
		Type:           s.p.Type,
	}, queryContext)
	if err != nil {
		if ierrors.ErrTimeout.Is(err) {
			return []*m.Event{}, cursor, nil
		}
		return nil, "", fmt.Errorf("failed to get events: %w", err)
	}
	return res.Events, res.ContinuationID, nil
}

func (s *EventSource) waitWithContext(dur time.Duration) error {
	select {
	case <-s.m.ctx.Done():
		return s.m.ctx.Err()
	case <-time.After(dur):
		return nil
	}
}

func (s *EventSource) Id() string {
	return s.p.Id
}
