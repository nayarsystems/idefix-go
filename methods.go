package idefixgo

import (
	"fmt"
	"time"

	"github.com/nayarsystems/pubsub-go/ps"
	"gitlab.com/garagemakers/idefix/core/idefix"
)

func (im *Client) Publish(remoteAddress string, msg *Message) error {
	msg.To = fmt.Sprintf("%s.%s", remoteAddress, msg.To)
	return im.sendMessage(msg)
}

func (im *Client) Call(remoteAddress string, msg *Message, timeout time.Duration) (*ps.Msg, error) {
	msg.To = fmt.Sprintf("%s.%s", remoteAddress, msg.To)
	msg.Response, _ = randSessionID()

	su := ps.NewSubscriber(1, msg.Response)
	im.sendMessage(msg)

	ret := su.GetWithCtx(im.ctx, timeout)
	if ret == nil {
		return nil, idefix.ErrInvalidParams
	}

	return ret, nil
}
