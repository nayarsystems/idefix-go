package eventpipe

import (
	"context"

	m "github.com/nayarsystems/idefix-go/messages"
)

// IdefixClient is an interface for the Idefix client methods used by the event pipeline
type IdefixClient interface {
	EventsGet(msg *m.EventsGetMsg, ctx ...context.Context) (*m.EventsGetResponseMsg, error)
}
