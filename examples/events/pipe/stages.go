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
}

func (s *mockStage) Process(ctx context.Context, input eventpipe.EventStageInput) (eventpipe.EventStageOutput, error) {
	contextData := input.PipelineContext
	index := ei.N(contextData).M("index").IntZ()
	if index == s.index {
		// Process the event

		ctx, cancel := context.WithTimeout(ctx, s.duration)
		defer cancel()
		<-ctx.Done()

		event := input.Event
		slog.Info(fmt.Sprintf("STAGE %d", s.index), "id", event.UID)

		// Update index in context
		index += 1
		contextData["index"] = index
	} else {
		slog.Warn("this cannot happen, skipping processing", "expected_index", s.index, "current_index", index)
	}

	out := eventpipe.EventStageOutput{
		Event:           input.Event,
		PipelineContext: contextData,
		Processed:       true,
	}
	return out, nil
}
