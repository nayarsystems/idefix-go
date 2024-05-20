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
			_, err := s.c.Call(s.address, &m.Message{To: m.TopicRemoteStartPublisher, Data: m.StreamCreateMsg{
				Id:          s.pubId,
				Timeout:     s.timeout,
				PayloadOnly: s.payloadOnly,
			}}, time.Second*5)
			if err != nil && !ie.ErrTimeout.Is(err) {
				s.cancel(err)
				return
			}
		}
	}
}

func (s *PublisherStream) Context() context.Context {
	return s.ctx
}

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
