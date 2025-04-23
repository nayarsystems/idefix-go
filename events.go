package idefixgo

import (
	"fmt"
	"time"

	m "github.com/nayarsystems/idefix-go/messages"
)

// SendEvent sends an event message with the specified payload, schema, metadata, and unique identifier (UID).
//
// The event message is constructed using the provided payload, which is an arbitrary interface{} type,
// and is associated with the schema defined by the hashSchema string. The meta parameter is a map
// containing additional metadata, and the uid is a string uniquely identifying the event.
//
// The event message is sent using MQTT to the "events.create" destination using the Idefix service.
// The call is made with a specified timeout, and if any error occurs during the process,
// the function returns the error.
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

// GetEvents retrieves a list of events from the specified domain and address that occurred after the given time.
//
// The function sends a request to the "events.get" destination using the Idefix service, and returns
// a response containing the events that match the provided criteria.
//
// The request includes a timeout duration, which is capped at 30 seconds if a larger value is provided.
// If successful, the function returns the retrieved events encapsulated in an EventsGetResponseMsg.
// If any error occurs, it returns an error.
//
// Deprecated: Use EventsGet instead.
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

// GetEvents retrieves an event from the specified UID.
//
// The function sends a request to the "events.get" destination using the Idefix service, and returns
// a response containing the events that match the provided criteria.
//
// The request includes a timeout duration, which is capped at 30 seconds if a larger value is provided.
// If successful, the function returns the retrieved events encapsulated in an EventsGetUIDResponseMsg.
// If any error occurs, it returns an error.
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
