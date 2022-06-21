package minips

import (
	"fmt"
	"regexp"
	"sync"
)

type Minips[T any] struct {
	m      sync.RWMutex
	topics map[string][]chan T
}

func NewMinips[T any]() *Minips[T] {
	mp := &Minips[T]{
		topics: make(map[string][]chan T),
	}

	return mp
}

func (mp *Minips[T]) RegisterChannel(topic string, ch chan T) {
	mp.m.Lock()
	defer mp.m.Unlock()

	list, ok := mp.topics[topic]
	if !ok {
		list = make([]chan T, 0)
	}

	mp.topics[topic] = append(list, ch)
}

func (mp *Minips[T]) UnregisterChannel(topic string, ch chan T) {
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

func (mp *Minips[T]) UnregisterChannelFromAll(ch chan T) {
	mp.m.RLock()
	defer mp.m.RUnlock()

	for topic, chans := range mp.topics {
		for _, ch2 := range chans {
			if ch2 == ch {
				mp.unregisterChannelSafe(topic, ch)
			}
		}
	}
}

func (mp *Minips[T]) Publish(topic string, elem T) {
	mp.m.Lock()
	defer mp.m.Unlock()

	mp.publishTopic(topic, elem)

	re := regexp.MustCompile(`^(.*)[.][^.]+$`)
	remainder := topic

	for {
		match := re.FindStringSubmatch(remainder)
		if match == nil {
			return
		}

		remainder = match[1]
		mp.publishTopic(remainder, elem)
	}
}

func (mp *Minips[T]) publishTopic(topic string, elem T) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered:", r)
		}
	}()

	list, ok := mp.topics[topic]
	if !ok {
		return
	}

	for _, v := range list {
		select {
		case v <- elem:
		default:
		}
	}
}
