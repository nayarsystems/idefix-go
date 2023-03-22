package idefixgo

import (
	"fmt"
	ie "github.com/nayarsystems/idefix-go/errors"
	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/nayarsystems/idefix-go/minips"
	"time"
)

func (c *Client) Publish(remoteAddress string, msg *m.Message) error {
	msg.To = fmt.Sprintf("%s.%s", remoteAddress, msg.To)
	return c.sendMessage(msg)
}

func (c *Client) Answer(origmsg *m.Message, msg *m.Message) error {
	msg.To = origmsg.Res // TODO: Check this
	return c.sendMessage(msg)
}

func (c *Client) Call(remoteAddress string, msg *m.Message, timeout time.Duration) (*m.Message, error) {
	var err error
	msg.To = fmt.Sprintf("%s.%s", remoteAddress, msg.To)
	msg.Res, err = randSessionID()
	if err != nil {
		return nil, err
	}

	sub := c.ps.NewSubscriber(1, msg.Res)
	defer sub.Close()

	if err := c.sendMessage(msg); err != nil {
		return nil, err
	}
	msg, err = sub.WaitOne(timeout)
	if err != nil {
		return nil, ie.ErrTimeout
	}
	if msg.Err != "" {
		return msg, fmt.Errorf(msg.Err)
	}
	return msg, nil
}

func (c *Client) Call2(remoteAddress string, msg *m.Message, resp any, timeout time.Duration) error {
	amap, err := m.ToMsi(msg.Data)
	if err != nil {
		return err
	}
	msg.Data = amap
	ret, err := c.Call(remoteAddress, msg, timeout)
	if err != nil {
		return err
	}
	if ret.Err != "" {
		return fmt.Errorf(msg.Err)
	}
	if resp != nil {
		respMsi, err := m.MsiCast(ret.Data)
		if err != nil {
			return err
		}
		return m.ParseMsi(respMsi, resp)
	}
	return nil
}

func (c *Client) NewSubscriber(capacity uint, topic ...string) *minips.Subscriber[*m.Message] {
	return c.ps.NewSubscriber(capacity, topic...)
}

func (c *Client) WaitOne(topic string, timeout time.Duration) (*m.Message, error) {
	sub := c.ps.NewSubscriber(1, topic)
	defer sub.Close()
	return sub.WaitOne(timeout)
}
