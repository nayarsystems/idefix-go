package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jaracil/ei"
	"github.com/nayarsystems/idefix-go/eventpipe"
)

type mockStage struct {
	duration time.Duration
	index    int
	remove   bool
}

func (s *mockStage) Process(ctx context.Context, input eventpipe.EventStageInput) (eventpipe.EventStageOutput, error) {
	contextData := input.PipelineContext
	if contextData == nil {
		contextData = map[string]any{}
	}

	index := ei.N(contextData).M("index").IntZ()
	if index == s.index {
		// Process the event
		event := input.Event
		slog.Info(fmt.Sprintf("STAGE %d", s.index), "id", event.UID)

		ctx, cancel := context.WithTimeout(ctx, s.duration)
		defer cancel()
		<-ctx.Done()

		// Update index in context
		index += 1
		contextData["index"] = index
	}

	out := eventpipe.EventStageOutput{
		Event:           input.Event,
		PipelineContext: contextData,
		Remove:          s.remove,
	}
	return out, nil
}
