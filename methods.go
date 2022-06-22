package idefixgo

import (
	"fmt"
	"time"

	"gitlab.com/garagemakers/idefix-go/minips"
)

func (im *Client) Publish(remoteAddress string, msg *Message) error {
	msg.To = fmt.Sprintf("%s.%s", remoteAddress, msg.To)
	return im.sendMessage(msg)
}

func (im *Client) Answer(origmsg *Message, msg *Message) error {
	msg.To = origmsg.Response // TODO: Check this
	return im.sendMessage(msg)
}

func (im *Client) Call(remoteAddress string, msg *Message, timeout time.Duration) (interface{}, error) {
	msg.To = fmt.Sprintf("%s.%s", remoteAddress, msg.To)
	msg.Response, _ = randSessionID()

	sub := im.ps.NewSubscriber(1, msg.Response)
	defer sub.Close()

	if err := im.sendMessage(msg); err != nil {
		return nil, err
	}

	return sub.WaitOne(timeout)
}

func (im *Client) NewSubscriber(capacity uint, topic ...string) *minips.Subscriber[*Message] {
	return im.ps.NewSubscriber(capacity, topic...)
}

func (im *Client) WaitOne(topic string, timeout time.Duration) (*Message, error) {
	sub := im.ps.NewSubscriber(1, topic)
	defer sub.Close()
	return sub.WaitOne(timeout)
}
