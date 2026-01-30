package eventpipe

import (
	"context"
	"errors"
	"fmt"

	m "github.com/nayarsystems/idefix-go/messages"
)

// Create custom errors
var (
	ErrUnableToProcessNow = errors.New("cannot process event now. Try later")
	ErrSkipEvent          = errors.New("skip this event")
)

type EventStage interface {
	Process(context.Context, EventStageInput) (EventStageOutput, error)
}

type EventStageInput struct {
	Event           *m.Event
	PipelineContext map[string]any
}

type EventStageOutput struct {
	Event           *m.Event
	PipelineContext map[string]any

	// If set to true, the event process will stop here (in this stage)
	// and the event will be removed from the pipeline.
	Remove bool

	// If Remove is true, this flag is ignored.
	// Marks the event as processed. Processed events will not be
	// retried again in the stages where the event was already processed.
	// This is useful when an event is re-injected into the pipeline
	// but we want to avoid re-processing it in the same stages.
	// If the last stage marks the event as Processed, the event
	// will be considered fully processed and removed from the pipeline as well.
	// This means that it is not necessary to set Remove=true in the last stage
	// if Processed=true is set.
	Processed bool
}

type EventStageOptionFn func(so *eventStageOptions) error

func OptName(name string) EventStageOptionFn {
	return func(so *eventStageOptions) error {
		so.name = name
		return nil
	}
}

func OptInputBufferSize(inputBufferSize uint) EventStageOptionFn {
	return func(so *eventStageOptions) error {
		if inputBufferSize == 0 {
			return fmt.Errorf("input buffer size must be at least 1")
		}
		so.bufferSize = inputBufferSize
		return nil
	}
}

func OptConcurrency(concurrency uint) EventStageOptionFn {
	return func(so *eventStageOptions) error {
		if concurrency == 0 {
			return fmt.Errorf("concurrency must be at least 1")
		}
		so.concurrency = concurrency
		return nil
	}
}
