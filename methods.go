package idefixgo

import (
	"fmt"
	"time"

	ie "github.com/nayarsystems/idefix-go/errors"
	"github.com/nayarsystems/idefix-go/minips"
)

func (c *Client) Publish(remoteAddress string, msg *Message) error {
	msg.To = fmt.Sprintf("%s.%s", remoteAddress, msg.To)
	return c.sendMessage(msg)
}

func (c *Client) Answer(origmsg *Message, msg *Message) error {
	msg.To = origmsg.Response // TODO: Check this
	return c.sendMessage(msg)
}

func (c *Client) Call(remoteAddress string, msg *Message, timeout time.Duration) (*Message, error) {
	var err error
	msg.To = fmt.Sprintf("%s.%s", remoteAddress, msg.To)
	msg.Response, err = randSessionID()
	if err != nil {
		return nil, err
	}

	sub := c.ps.NewSubscriber(1, msg.Response)
	defer sub.Close()

	if err := c.sendMessage(msg); err != nil {
		return nil, err
	}
	msg, err = sub.WaitOne(timeout)
	if err != nil {
		return nil, ie.ErrTimeout
	}
	if msg.Err != nil {
		return msg, msg.Err
	}
	return msg, nil
}

func (c *Client) NewSubscriber(capacity uint, topic ...string) *minips.Subscriber[*Message] {
	return c.ps.NewSubscriber(capacity, topic...)
}

func (c *Client) WaitOne(topic string, timeout time.Duration) (*Message, error) {
	sub := c.ps.NewSubscriber(1, topic)
	defer sub.Close()
	return sub.WaitOne(timeout)
}
