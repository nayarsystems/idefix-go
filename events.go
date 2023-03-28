package idefixgo

import (
	"fmt"
	"time"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/nayarsystems/idefix/libraries/eval"
)

type GetEventsBaseParams struct {
	Domain        string
	Limit         uint
	Cid           string
	Timeout       time.Duration
	Since         time.Time
	AddressFilter string
	MetaFilter    eval.CompiledExpr
	Continue      bool
	Csvdir        string
}

func (c *Client) SendEvent(payload interface{}, hashSchema string, meta map[string]interface{}, uid string, timeout time.Duration) error {
	msg := m.EventMsg{
		Type:    fmt.Sprintf(`application/vnd.nayar.bstates; id="%s"`, hashSchema),
		Payload: payload,
		Meta:    meta,
		UID:     uid,
	}
	err := c.Call2("idefix", &m.Message{To: "events.create", Data: &msg}, nil, timeout)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetEventsByDomain(domain string, since time.Time, limit uint, cid string, timeout time.Duration) (*m.EventsGetResponseMsg, error) {
	msg := m.EventsGetMsg{
		Domain:         domain,
		Since:          since,
		Limit:          limit,
		ContinuationID: cid,
	}
	if timeout > 0 {
		if timeout > time.Second*30 {
			timeout = time.Second * 30
		}

		msg.Timeout = timeout
	}
	resp := &m.EventsGetResponseMsg{}
	err := c.Call2("idefix", &m.Message{To: "events.get", Data: &msg}, resp, timeout+time.Second)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetSchema(hash string, timeout time.Duration) (*m.SchemaGetResponseMsg, error) {
	msg := m.SchemaGetMsg{
		Hash:  hash,
		Check: false,
	}
	resp := &m.SchemaGetResponseMsg{}
	err := c.Call2("idefix", &m.Message{To: "schemas.get", Data: &msg}, resp, timeout)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
