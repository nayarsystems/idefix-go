package eventpipe

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/stretchr/testify/require"
)

// --- Mock IdefixClient ---

type mockIdefixClient struct {
	events []*m.Event
}

func (c *mockIdefixClient) EventsGet(msg *m.EventsGetMsg, ctx ...context.Context) (*m.EventsGetResponseMsg, error) {
	if msg.ContinuationID == "" {
		return &m.EventsGetResponseMsg{
			Events:         c.events,
			ContinuationID: "cursor-after-batch",
		}, nil
	}

	// Already have a cursor, no new events
	return &m.EventsGetResponseMsg{
		Events:         []*m.Event{},
		ContinuationID: msg.ContinuationID,
	}, nil
}

// --- Mock EventStage ---

type mockStage struct {
	mu        sync.Mutex
	processed []string
}

func (s *mockStage) Process(ctx context.Context, input EventStageInput) (EventStageOutput, error) {
	s.mu.Lock()
	s.processed = append(s.processed, input.Event.UID)
	s.mu.Unlock()

	return EventStageOutput{
		Event:           input.Event,
		PipelineContext: input.PipelineContext,
		Processed:       true,
	}, nil
}

func (s *mockStage) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.processed)
}

// mockRetryStage returns Processed=false for the first `failCount` calls per event,
// then Processed=true. Tracks total calls and processed (successful) calls.
type mockRetryStage struct {
	mu            sync.Mutex
	failCount     int            // how many times to return Processed=false per event
	attempts      map[string]int // UID -> number of calls
	processedUIDs []string       // UIDs that were finally processed (Processed=true)
	totalCalls    int
}

func newMockRetryStage(failCount int) *mockRetryStage {
	return &mockRetryStage{
		failCount: failCount,
		attempts:  make(map[string]int),
	}
}

func (s *mockRetryStage) Process(ctx context.Context, input EventStageInput) (EventStageOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	uid := input.Event.UID
	s.attempts[uid]++
	s.totalCalls++

	processed := s.attempts[uid] > s.failCount
	if processed {
		s.processedUIDs = append(s.processedUIDs, uid)
	}

	return EventStageOutput{
		Event:           input.Event,
		PipelineContext: input.PipelineContext,
		Processed:       processed,
	}, nil
}

func (s *mockRetryStage) processedCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.processedUIDs)
}

func (s *mockRetryStage) totalCallCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.totalCalls
}

// --- Helpers ---

func runSource(t *testing.T, cancel context.CancelFunc, source *EventSource) {
	t.Helper()
	done := make(chan error, 1)
	go func() {
		done <- source.Run()
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("pipeline goroutine did not finish in time")
		}
	})
}

func generateTestEvents(n int) []*m.Event {
	events := make([]*m.Event, n)
	for i := range n {
		events[i] = &m.Event{
			EventMsg: m.EventMsg{
				UID:     fmt.Sprintf("event-%d", i),
				Type:    "test",
				Payload: fmt.Sprintf("payload-%d", i),
			},
			Domain:    "test-domain",
			Address:   "test-device",
			Timestamp: time.Now(),
		}
	}
	return events
}

// --- Tests ---

type storagePathFn func(t *testing.T) string

func storagePaths() map[string]storagePathFn {
	return map[string]storagePathFn{
		"memory": func(t *testing.T) string { return ":memory" },
		"sqlite": func(t *testing.T) string { return filepath.Join(t.TempDir(), "test.db") },
	}
}

func TestPipeline(t *testing.T) {
	for name, pathFn := range storagePaths() {
		t.Run(name, func(t *testing.T) {
			testHappyPath(t, pathFn(t))
			testProcessedFalseRetry(t, pathFn(t))
		})
	}
}

func testHappyPath(t *testing.T, storagePath string) {
	ctx, cancel := context.WithCancel(context.Background())

	mockClient := &mockIdefixClient{events: generateTestEvents(3)}

	esm, err := NewEventSourceManager(EventSourceManagerParams{
		Client:      mockClient,
		Context:     ctx,
		StoragePath: storagePath,
	})
	require.NoError(t, err)
	require.NoError(t, esm.Init())
	t.Cleanup(func() { esm.Close() })

	source, err := esm.NewSource(EventSourceParams{
		Id:     "test-source",
		Domain: "test-domain",
	})
	require.NoError(t, err)

	stage1 := &mockStage{}
	stage2 := &mockStage{}
	require.NoError(t, source.Push(stage1, OptName("stage-1")))
	require.NoError(t, source.Push(stage2, OptName("stage-2")))

	runSource(t, cancel, source)

	// Wait for both stages to process all 3 events
	require.Eventually(t, func() bool {
		return stage1.count() == 3 && stage2.count() == 3
	}, 30*time.Second, 100*time.Millisecond)

	// Verify storage is empty (all events fully processed and removed)
	events, err := (&EventsStorage{St: esm.st}).GetEvents("test-source")
	require.NoError(t, err)
	require.Empty(t, events)
}

func testProcessedFalseRetry(t *testing.T, storagePath string) {
	ctx, cancel := context.WithCancel(context.Background())

	mockClient := &mockIdefixClient{events: generateTestEvents(2)}

	esm, err := NewEventSourceManager(EventSourceManagerParams{
		Client:      mockClient,
		Context:     ctx,
		StoragePath: storagePath,
	})
	require.NoError(t, err)
	require.NoError(t, esm.Init())
	t.Cleanup(func() { esm.Close() })

	source, err := esm.NewSource(EventSourceParams{
		Id:     "test-source-retry",
		Domain: "test-domain",
	})
	require.NoError(t, err)

	// Stage 1: fails 2 times per event before succeeding
	retryStage := newMockRetryStage(2)
	// Stage 2: always succeeds
	finalStage := &mockStage{}

	require.NoError(t, source.Push(retryStage, OptName("retry-stage")))
	require.NoError(t, source.Push(finalStage, OptName("final-stage")))

	runSource(t, cancel, source)

	// Wait for both events to pass through both stages
	require.Eventually(t, func() bool {
		return retryStage.processedCount() == 2 && finalStage.count() == 2
	}, 30*time.Second, 100*time.Millisecond)

	// retryStage should have been called 3 times per event (2 fails + 1 success)
	require.Equal(t, 6, retryStage.totalCallCount())

	// Verify storage is empty
	events, err := (&EventsStorage{St: esm.st}).GetEvents("test-source-retry")
	require.NoError(t, err)
	require.Empty(t, events)
}
