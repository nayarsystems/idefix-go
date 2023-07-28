package idefixgo

import (
	"time"

	m "github.com/nayarsystems/idefix-go/messages"
)

func (c *Client) GetAddressDomain(address string, timeout time.Duration) (*m.Domain, error) {
	msg := m.AddressDomainGetMsg{
		Address: address,
	}
	resp := &m.Domain{}
	err := c.Call2("idefix", &m.Message{To: m.CmdAddressDomainGet, Data: &msg}, resp, timeout)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
