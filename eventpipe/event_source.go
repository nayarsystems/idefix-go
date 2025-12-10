package eventpipe

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/go-pipeline/pkg/pipeline"
	ierrors "github.com/nayarsystems/idefix-go/errors"
	m "github.com/nayarsystems/idefix-go/messages"
)

type EventSource struct {
	l      *slog.Logger
	m      *EventSourceManager
	p      EventSourceParams
	stages []pipeline.Stage[pipelineItem]
}

type EventSourceParams struct {
	// Query events since this time if no cursor found in storage
	// If zero, query all events (limited by service query limits)
	Since time.Time

	// // Whether to restart cursor from RestartCursorFrom time
	// RestartCursor bool

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
	// if eventStageOptions.name == "" {
	// 	return fmt.Errorf("stage name is required")
	// }
	gopts = append(gopts, pipeline.Name(eventStageOptions.name))
	gopts = append(gopts, pipeline.Concurrency(eventStageOptions.concurrency))
	gopts = append(gopts, pipeline.InputBufferSize(eventStageOptions.bufferSize))

	gstage := pipeline.NewStage(
		func(in pipelineItem) (out pipelineItem, err error) {
			if s.m.ctx.Err() != nil {
				return in, s.m.ctx.Err()
			}
			if in.passthrough {
				return in, nil
			}
			stageInput := EventStageInput{
				Event:           in.event,
				PipelineContext: in.eventContext,
			}
			var stageOutput EventStageOutput
			stageOutput, err = stage.Process(s.m.ctx, stageInput)
			if err != nil {
				return out, err
			}
			if stageOutput.Remove {
				s.deleteEvent(in.event)
				out.passthrough = true
				return out, nil
			}
			out.event = stageOutput.Event
			out.eventContext = stageOutput.PipelineContext
			if err = s.updateEvent(out.event, out.eventContext); err != nil {
				return out, err
			}
			return out, nil
		},
		gopts...,
	)
	s.stages = append(s.stages, gstage)
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
	var events []*m.Event
	cursor, err := s.m.st.GetCursor(s.Id())
	if err != nil {
		return fmt.Errorf("failed to get cursor: %w", err)
	}
	for s.waitWithContext(time.Second) == nil {
		storageEvents, err := s.loadEvents()
		if err != nil {
			return fmt.Errorf("failed to load pending events: %w", err)
		}
		for _, e := range storageEvents {
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

func (s *EventSource) loadEvents() ([]*eventItem, error) {
	db := EventsStorage{St: s.m.st}
	events, err := db.GetEvents(s.Id())
	if err != nil {
		return nil, fmt.Errorf("failed to load pending events: %w", err)
	}
	return events, nil
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
