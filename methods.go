package idefixgo

import (
	"fmt"
	"time"

	ie "github.com/nayarsystems/idefix-go/errors"
	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/nayarsystems/idefix-go/minips"
)

// Publish sends a message to a specified remote address.
func (c *Client) Publish(remoteAddress string, msg *m.Message) error {
	msg.To = fmt.Sprintf("%s.%s", remoteAddress, msg.To)
	return c.sendMessage(msg)
}

// Answer constructs a response message based on the original message
// and sends it back to the intended recipient.
//
// The original message, 'origmsg', is used to determine the response
// destination by setting the 'To' field of the response message `msg`
// to the 'Res' field of the original message. This allows the sender
// of the original message to receive the response correctly.
//
// It is important to ensure that 'origmsg' contains a valid response
// destination in its 'Res' field before calling this method. If the
// response message fails to send, an error will be returned.
func (c *Client) Answer(origmsg *m.Message, msg *m.Message) error {
	msg.To = origmsg.Res // TODO: Check this
	return c.sendMessage(msg)
}

// Call sends a message to a specified remote address and expects a response. If timeout given is exceed it returns an error.
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

// Call2 uses [Client.Call] to send a message to a specified remote address and expects a response.
// The function converts the 'msg.Data' field into a map (msi) format before sending it using the internal [Client.Call] method.
// If a response is expected ('resp' is not nil), it parses the returned data into the provided response structure.
// The function also handles any errors that occur during the process, either in the message sending or the response.
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
		respMsi, err := m.GetMsi(ret.Data)
		if err != nil {
			return err
		}
		return m.ParseMsi(respMsi, resp)
	}
	return nil
}

// NewSubscriber creates a new message subscriber with the specified capacity and topic(s).
//
// The function initializes a new subscriber using the client's internal publish-subscribe system ('ps'),
// allowing it to receive messages published on the specified topics. The subscriber is configured with a given
// buffer capacity to manage the number of messages it can hold before processing.
func (c *Client) NewSubscriber(capacity uint, topic ...string) *minips.Subscriber[*m.Message] {
	return c.ps.NewSubscriber(capacity, topic...)
}

// WaitOne waits for a single message on the specified topic within the given timeout duration.
//
// The function subscribes to the specified topic and blocks until a message is received or the timeout occurs.
// If a message is received within the timeout, it is returned. If the timeout expires before a message arrives,
// an error is returned.
func (c *Client) WaitOne(topic string, timeout time.Duration) (*m.Message, error) {
	sub := c.ps.NewSubscriber(1, topic)
	defer sub.Close()
	return sub.WaitOne(timeout)
}
