package idefixgo

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/garagemakers/idefix-go/minips"
)

type Stream struct {
	ctx     context.Context
	cancel  context.CancelFunc
	timeout time.Duration
	c       *Client
	sub     *minips.Subscriber[*Message]
	topic   string
	address string
}

func (c *Client) NewStream(address string, topic string, capacity uint, timeout time.Duration) (*Stream, error) {
	s := &Stream{
		address: address,
		timeout: timeout,
		topic:   topic,
		c:       c,
		sub:     c.ps.NewSubscriber(capacity),
	}

	if err := s.sub.Subscribe(s.subTopic()); err != nil {
		return nil, err
	}

	s.ctx, s.cancel = context.WithCancel(c.ctx)

	_, err := s.c.Call(address, &Message{To: s.openTopic(), Data: map[string]any{"timeout": s.timeout.Seconds()}}, time.Second*5)
	if err != nil {
		return nil, err
	}

	go s.keepalive()

	return s, nil
}

func (s *Stream) openTopic() string {
	return "$ss." + s.topic
}

func (s *Stream) subTopic() string {
	return fmt.Sprintf("$sm.%s.%s", s.address, s.topic)
}

func (s *Stream) closeTopic() string {
	return "$su." + s.topic
}

func (s *Stream) keepalive() {
	t := time.NewTicker(s.timeout / 2)
	defer s.Close()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-t.C:
			_, err := s.c.Call(s.address, &Message{To: s.openTopic(), Data: map[string]any{"timeout": s.timeout.Seconds()}}, time.Second*5)
			if err != nil {
				return
			}
		}
	}
}

func (s *Stream) Channel() <-chan *Message {
	return s.sub.Channel()
}

func (s *Stream) Close() error {
	defer s.cancel()
	_, err := s.c.Call(s.address, &Message{To: s.closeTopic()}, s.timeout)
	if err != nil {
		return err
	}
	return nil
}
