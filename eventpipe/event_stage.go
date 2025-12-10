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
	Remove          bool
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
