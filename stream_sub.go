package idefixgo

import (
	"context"
	"fmt"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jaracil/ei"
	ie "github.com/nayarsystems/idefix-go/errors"
	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/vmihailenco/msgpack/v5"
)

type SubscriberStream struct {
	ctx         context.Context
	cancel      context.CancelCauseFunc
	timeout     time.Duration
	c           *Client
	buffer      chan *m.Message
	topic       string
	address     string
	subId       string
	payloadOnly bool
}

func (c *Client) NewSubscriberStream(address string, topic string, capacity uint, payloadOnly bool, timeout time.Duration) (*SubscriberStream, error) {
	s := &SubscriberStream{
		address:     address,
		timeout:     timeout,
		topic:       topic,
		c:           c,
		buffer:      make(chan *m.Message, capacity),
		payloadOnly: payloadOnly,
	}

	s.ctx, s.cancel = context.WithCancelCause(c.ctx)

	res := m.StreamCreateSubResMsg{}
	err := s.c.Call2(address, &m.Message{To: m.TopicRemoteSubscribe, Data: m.StreamCreateMsg{
		TargetTopic: topic,
		Timeout:     time.Second * 30,
		PayloadOnly: s.payloadOnly,
	}}, &res, time.Second*5)
	if err != nil {
		return nil, err
	}

	if res.StickyPayload != nil {
		s.handleMsg(res.StickyPayload)
	}

	pubtopic := fmt.Sprintf("%s/%s", m.MqttPublicPrefix, res.PublicTopic)

	c.client.Subscribe(pubtopic, 2, s.receiveMessage)

	s.subId = res.Id
	go s.keepalive()

	return s, nil
}

func (s *SubscriberStream) handleMsg(msg any) {
	if s.payloadOnly {
		s.buffer <- &m.Message{To: s.topic, Data: msg}
		return
	}

	topic, err := ei.N(msg).M("s").String()
	if err != nil {
		topic = s.topic
	}

	payload, err := ei.N(msg).M("p").Raw()
	if err != nil {
		fmt.Println("Error getting payload", err)
		return
	}

	s.buffer <- &m.Message{To: topic, Data: payload}
}

func (s *SubscriberStream) receiveMessage(client mqtt.Client, msg mqtt.Message) {
	if strings.HasPrefix(msg.Topic(), m.MqttPublicPrefix+"/") {
		var tmp any
		err := msgpack.Unmarshal(msg.Payload(), &tmp)
		if err != nil {
			fmt.Println("Error unmarshalling message", err, msg.Payload())
			return
		}
		s.handleMsg(tmp)
	}
}

func (s *SubscriberStream) keepalive() {
	t := time.NewTicker(s.timeout / 4)
	defer s.Close()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-t.C:
			_, err := s.c.Call(s.address, &m.Message{To: m.TopicRemoteSubscribe, Data: m.StreamCreateMsg{
				Id:          s.subId,
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

func (s *SubscriberStream) Channel() <-chan *m.Message {
	return s.buffer
}

func (s *SubscriberStream) Context() context.Context {
	return s.ctx
}

func (s *SubscriberStream) Close() error {
	defer s.cancel(fmt.Errorf("closed by user"))
	_, err := s.c.Call(s.address, &m.Message{To: m.TopicRemoteUnsubscribe, Data: m.StreamDeleteMsg{
		Id: s.subId,
	}}, time.Second*5)
	if err != nil {
		return err
	}
	return nil
}
