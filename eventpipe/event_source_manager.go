package eventpipe

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nayarsystems/idefix-go/eventpipe/storage"
)

type EventSourceManager struct {
	l      *slog.Logger
	p      EventSourceManagerParams
	c      IdefixClient
	ctx    context.Context
	cancel context.CancelCauseFunc
	st     storage.Storage
}

type EventSourceManagerParams struct {
	Logger      *slog.Logger
	Client      IdefixClient
	Context     context.Context
	StoragePath string // path to storage directory
}

func NewEventSourceManager(params EventSourceManagerParams) (*EventSourceManager, error) {
	esm := &EventSourceManager{p: params, c: params.Client}
	if params.Logger != nil {
		esm.l = params.Logger
	} else {
		esm.l = slog.Default()
	}

	ctx := context.Background()
	if params.Context != nil {
		ctx = params.Context
	}
	esm.ctx, esm.cancel = context.WithCancelCause(ctx)

	// Choose database type based on StoragePath
	dbType := "sqlite"
	if params.StoragePath == ":memory" {
		dbType = "memory"
	}

	esm.st = storage.NewStorage(dbType)
	return esm, nil
}

func (m *EventSourceManager) Init() error {
	if err := m.st.Init(m.ctx, map[string]any{
		"path": m.p.StoragePath,
	}); err != nil {
		return err
	}
	return nil
}

func (m *EventSourceManager) Context() context.Context {
	return m.ctx
}

func (m *EventSourceManager) Close() error {
	m.cancel(nil)
	if err := m.st.Close(); err != nil {
		m.l.Error("failed to close database", "error", err)
		return err
	}
	return nil
}

func (m *EventSourceManager) NewSource(params EventSourceParams) (*EventSource, error) {
	es := &EventSource{m: m}
	if params.Id == "" {
		return nil, fmt.Errorf("event source id must be specified")
	}
	if params.Domain == "" && params.Address == "" {
		return nil, fmt.Errorf("either domain or address must be specified for event source %s", params.Id)
	}
	if params.LongPollingTimeout == 0 {
		params.LongPollingTimeout = time.Minute
	}
	es.l = m.l.With(slog.String("domain", params.Domain))
	if params.Address != "" {
		es.l = es.l.With(slog.String("address", params.Address))
	}
	es.l = es.l.With(slog.String("sourceId", es.Id()))
	es.p = params
	return es, nil
}
