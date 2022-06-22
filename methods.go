package idefixgo

import (
	"context"
	"fmt"
	"time"
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

	ch := make(chan *Message, 1)
	im.ps.RegisterChannel(msg.Response, ch)

	defer close(ch)
	defer im.ps.UnregisterChannel(msg.Response, ch)

	if err := im.sendMessage(msg); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithDeadline(im.ctx, time.Now().Add(timeout))
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ErrTimeout
		case m, ok := <-ch:
			if !ok {
				return nil, ErrChannelClosed
			}
			if m.To == msg.Response {
				return m.Data, m.Err
			}
		}
	}
}

func (im *Client) Subscribe(topic string, ch chan *Message) {
	im.ps.RegisterChannel(topic, ch)
	return
}

func (im *Client) Unsubscribe(topic string, ch chan *Message) {
	im.ps.UnregisterChannel(topic, ch)
	return
}

func (im *Client) UnsubscribeAll(ch chan *Message) {
	im.ps.UnregisterChannelFromAll(ch)
	return
}

func (im *Client) WaitOne(topic string, timeout time.Duration) (*Message, error) {
	ch := make(chan *Message, 1)
	im.ps.RegisterChannel(topic, ch)

	defer close(ch)
	defer im.ps.UnregisterChannel(topic, ch)

	ctx, cancel := context.WithDeadline(im.ctx, time.Now().Add(timeout))
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ErrTimeout
		case m, ok := <-ch:
			if !ok {
				return nil, ErrChannelClosed
			}
			return m, nil
		}
	}
}
