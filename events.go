package idefixgo

import (
	"encoding/json"
	"time"
)

func (c *Client) SendEvent(payload interface{}, hashSchema string, meta map[string]interface{}, timeout time.Duration) error {
	amap := make(map[string]interface{})
	amap["payload"] = payload
	amap["schema"] = hashSchema
	amap["meta"] = meta

	ret, err := c.Call("idefix", &Message{To: "events.create", Data: amap}, timeout)
	if err != nil {
		return err
	}

	if ret.Err != nil {
		return ret.Err
	}

	return nil
}

func (c *Client) GetEventsByDomain(domain string, since time.Time, limit uint, skip uint, reverse bool, timeout time.Duration) ([]Event, error) {
	amap := make(map[string]interface{})
	amap["domain"] = domain
	amap["since"] = since
	amap["limit"] = limit
	amap["skip"] = skip
	amap["reverse"] = reverse

	ret, err := c.Call("idefix", &Message{To: "events.get", Data: amap}, timeout)
	if err != nil {
		return nil, err
	}

	if ret.Err != nil {
		return nil, ret.Err
	}

	b, err := json.Marshal(ret.Data)
	if err != nil {
		return nil, err
	}
	m := []Event{}
	_ = json.Unmarshal(b, &m)

	return m, nil
}

func (c *Client) GetSchema(hash string, timeout time.Duration) (*Schema, error) {
	amap := make(map[string]interface{})
	amap["hash"] = hash
	amap["nopayload"] = false

	ret, err := c.Call("idefix", &Message{To: "schemas.get", Data: amap}, timeout)
	if err != nil {
		return nil, err
	}

	if ret.Err != nil {
		return nil, ret.Err
	}

	b, err := json.Marshal(ret.Data)
	if err != nil {
		return nil, err
	}
	m := &Schema{}
	_ = json.Unmarshal(b, &m)

	return m, nil
}
