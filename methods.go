package idefixgo

import (
	"fmt"
	"time"

	"gitlab.com/garagemakers/idefix-go/minips"
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
	msg.To = fmt.Sprintf("%s.%s", remoteAddress, msg.To)
	msg.Response, _ = randSessionID()

	sub := c.ps.NewSubscriber(1, msg.Response)
	defer sub.Close()

	if err := c.sendMessage(msg); err != nil {
		return nil, err
	}

	return sub.WaitOne(timeout)
}

func (c *Client) NewSubscriber(capacity uint, topic ...string) *minips.Subscriber[*Message] {
	return c.ps.NewSubscriber(capacity, topic...)
}

func (c *Client) WaitOne(topic string, timeout time.Duration) (*Message, error) {
	sub := c.ps.NewSubscriber(1, topic)
	defer sub.Close()
	return sub.WaitOne(timeout)
}
