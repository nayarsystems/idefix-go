package minips

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"
)

type Minips[T any] struct {
	ctx    context.Context
	cancel context.CancelFunc
	m      sync.RWMutex
	topics map[string][]chan T
}

type Subscriber[T any] struct {
	ch     chan T
	mp     *Minips[T]
	closed bool
}

func (mp *Minips[T]) NewSubscriber(capacity uint, topics ...string) *Subscriber[T] {
	s := &Subscriber[T]{}
	s.mp = mp
	s.ch = make(chan T, capacity)

	for _, topic := range topics {
		s.Subscribe(topic)
	}
	return s
}

func (s *Subscriber[T]) Channel() <-chan T {
	return s.ch
}

func (s *Subscriber[T]) WaitOne(timeout time.Duration) (T, error) {
	ctx, cancel := context.WithDeadline(s.mp.ctx, time.Now().Add(timeout))
	defer cancel()

	var zeroT T
	for {
		select {
		case <-ctx.Done():
			return zeroT, fmt.Errorf("timeout")

		case m, ok := <-s.ch:
			if !ok {
				return zeroT, fmt.Errorf("closed")
			}
			return m, nil
		}
	}
}

func (s *Subscriber[T]) Subscribe(topics ...string) error {
	if s.closed {
		return fmt.Errorf("closed")
	}

	for _, topic := range topics {
		s.mp.registerChannel(topic, s.ch)
	}
	return nil
}

func (s *Subscriber[T]) Unsubscribe(topics ...string) {
	for _, topic := range topics {
		s.mp.unregisterChannel(topic, s.ch)
	}
}

func (s *Subscriber[T]) UnsubscribeAll() {
	s.mp.unregisterChannelFromAll(s.ch)
}

func (s *Subscriber[T]) Close() {
	s.closed = true
	s.mp.unregisterChannelFromAll(s.ch)
	close(s.ch)
}

func NewMinips[T any](pctx context.Context) *Minips[T] {
	mp := &Minips[T]{
		topics: make(map[string][]chan T),
	}
	mp.ctx, mp.cancel = context.WithCancel(pctx)

	return mp
}

func (mp *Minips[T]) registerChannel(topic string, ch chan T) {
	mp.m.Lock()
	defer mp.m.Unlock()

	list, ok := mp.topics[topic]
	if !ok {
		list = make([]chan T, 0)
	}

	mp.topics[topic] = append(list, ch)
}

func (mp *Minips[T]) unregisterChannel(topic string, ch chan T) {
	mp.m.Lock()
	defer mp.m.Unlock()

	mp.unregisterChannelSafe(topic, ch)
}

func (mp *Minips[T]) unregisterChannelSafe(topic string, ch chan T) {
	list, ok := mp.topics[topic]
	if !ok {
		return
	}

	for k, v := range list {
		if v == ch {
			mp.topics[topic] = append(list[:k], list[k+1:]...)
		}
	}

	if len(mp.topics[topic]) == 0 {
		delete(mp.topics, topic)
	}
}

func (mp *Minips[T]) unregisterChannelFromAll(ch chan T) {
	mp.m.Lock()
	defer mp.m.Unlock()

	for topic, chans := range mp.topics {
		for _, ch2 := range chans {
			if ch2 == ch {
				mp.unregisterChannelSafe(topic, ch)
			}
		}
	}
}

var topicSplitterRegexp = regexp.MustCompile(`^(.*)[.][^.]+$`)

func (mp *Minips[T]) Publish(topic string, elem T) uint {
	mp.m.Lock()
	defer mp.m.Unlock()
	var receivers uint

	receivers += mp.publishTopic(topic, elem)

	remainder := topic

	for {
		match := topicSplitterRegexp.FindStringSubmatch(remainder)
		if match == nil {
			return receivers
		}

		remainder = match[1]
		receivers += mp.publishTopic(remainder, elem)
	}
}

func (mp *Minips[T]) publishTopic(topic string, elem T) uint {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("panic publishing on topic %s: %s\n", topic, r)
		}
	}()

	list, ok := mp.topics[topic]
	if !ok {
		return 0
	}

	var receivers uint

	for _, v := range list {
		select {
		case v <- elem:
			receivers++
		default:
		}
	}

	return receivers
}

func (mp *Minips[T]) Close() {
	mp.cancel()
}
