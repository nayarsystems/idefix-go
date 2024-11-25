package idefixgo

import (
	"context"
	"fmt"
	"time"

	ie "github.com/nayarsystems/idefix-go/errors"
	"github.com/nayarsystems/idefix-go/messages"
	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/vmihailenco/msgpack/v5"
)

// PublisherStream represents a stream for publishing messages to a specific topic.
// It manages the context, timeout, and associated client details for effective message
// publishing.
type PublisherStream struct {
	ctx         context.Context
	cancel      context.CancelCauseFunc
	timeout     time.Duration
	c           *Client
	topic       string
	address     string
	pubId       string
	payloadOnly bool
	publicTopic string
}

// NewPublisherStream creates a new PublisherStream instance for publishing messages
// to a specified topic on the remote system. It initializes the necessary context,
// establishes the publisher connection, and configures the stream based on provided
// parameters.
//
// This function connects to the specified address, sets up the necessary context,
// and sends a request to start publishing on the specified topic. It also manages
// the lifetime of the PublisherStream through context cancellation
func (c *Client) NewPublisherStream(address string, topic string, capacity uint, payloadOnly bool, timeout time.Duration) (*PublisherStream, error) {
	s := &PublisherStream{
		address:     address,
		timeout:     timeout,
		topic:       topic,
		c:           c,
		payloadOnly: payloadOnly,
	}

	s.ctx, s.cancel = context.WithCancelCause(c.ctx)

	res := m.StreamCreateSubResMsg{}
	err := s.c.Call2(address, &m.Message{To: m.TopicRemoteStartPublisher, Data: m.StreamCreateMsg{
		TargetTopic: topic,
		Timeout:     time.Second * 30,
		PayloadOnly: s.payloadOnly,
	}}, &res, time.Second*5)
	if err != nil {
		return nil, err
	}

	s.pubId = res.Id
	go s.keepalive()

	s.publicTopic = fmt.Sprintf("%s/%s", m.MqttPublicPrefix, res.PublicTopic)

	return s, nil
}

// Publish sends a message to a specific subtopic of the PublisherStream's main topic.
//
// The method determines whether to publish only the payload or to wrap it in a
// StreamMsg structure based on the payloadOnly flag. If payloadOnly is true,
// the message is sent directly; otherwise, it is encapsulated within a StreamMsg.
func (s *PublisherStream) Publish(msg any, subtopic string) error {
	targetTopic := fmt.Sprintf("%s.%s", s.topic, subtopic)

	var payload any
	if s.payloadOnly {
		payload = msg
	} else {
		payload = messages.StreamMsg{
			SourceTopic: targetTopic,
			Payload:     msg,
		}
	}
	mqttPayload, err := msgpack.Marshal(payload)
	if err != nil {
		return err
	}
	token := s.c.client.Publish(s.publicTopic, 0, false, mqttPayload)
	token.Wait()
	return token.Error()

}

func (s *PublisherStream) keepalive() {
	t := time.NewTicker(s.timeout / 4)
	defer s.Close()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-t.C:
			err := s.c.Call2(s.address, &m.Message{To: m.TopicRemoteStartPublisher, Data: m.StreamCreateMsg{
				Id:          s.pubId,
				Timeout:     s.timeout,
				PayloadOnly: s.payloadOnly,
			}}, nil, time.Second*5)
			if err != nil && !ie.ErrTimeout.Is(err) {
				s.cancel(err)
				return
			}
		}
	}
}

// Context returns the context associated with the PublisherStream.
// This context can be used to manage the lifecycle of the stream,
// allowing for cancellation and timeout control.
func (s *PublisherStream) Context() context.Context {
	return s.ctx
}

// Close terminates the PublisherStream, releasing any associated resources.
// It cancels the context of the stream and sends a request to the remote
// system to stop the publisher associated with this stream.
//
// The method sends a message to the specified address with the command
// to stop the publisher identified by the stream's ID. It waits for a
// response, timing out after five seconds if no response is received.
func (s *PublisherStream) Close() error {
	defer s.cancel(fmt.Errorf("closed by user"))
	_, err := s.c.Call(s.address, &m.Message{To: m.TopicRemoteStopPublisher, Data: m.StreamDeleteMsg{
		Id: s.pubId,
	}}, time.Second*5)
	if err != nil {
		return err
	}
	return nil
}
