package idefixgo

import (
	"fmt"
	"time"

	m "github.com/nayarsystems/idefix-go/messages"
)

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

func (c *Client) GetEvents(domain, address string, since time.Time, limit uint, cid string, timeout time.Duration) (*m.EventsGetResponseMsg, error) {
	msg := m.EventsGetMsg{
		Address:        address,
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

func (c *Client) GetEventByUID(uid string, timeout time.Duration) (*m.EventsGetUIDResponseMsg, error) {
	msg := m.EventsGetMsg{
		UID: uid,
	}
	if timeout > 0 {
		if timeout > time.Second*30 {
			timeout = time.Second * 30
		}

		msg.Timeout = timeout
	}
	resp := &m.EventsGetUIDResponseMsg{}
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
