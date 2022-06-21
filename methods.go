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
