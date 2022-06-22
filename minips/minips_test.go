package minips

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMinips(t *testing.T) {
	mp := NewMinips[string](context.Background())

	ch1 := make(chan string, 100)
	ch11 := make(chan string, 100)
	ch2 := make(chan string, 100)

	mp.registerChannel("test1", ch1)
	mp.registerChannel("test2", ch2)
	mp.registerChannel("test1", ch11)

	np := mp.Publish("test1.asdf.qwe", "hola")
	require.Equal(t, uint(2), np)

	mp.Publish("test1.asdf", "hola")
	mp.Publish("test1", "hola")
	np = mp.Publish("test2", "hola")
	require.Equal(t, uint(1), np)

	require.Equal(t, 3, len(ch1))
	require.Equal(t, 3, len(ch11))
	require.Equal(t, 1, len(ch2))

	mp.unregisterChannel("test2", ch2)
	np = mp.Publish("test2", "hola")
	require.Equal(t, uint(0), np)
	require.Equal(t, 1, len(ch2))

	close(ch2)
	mp.registerChannel("test2", ch2)
	mp.Publish("test2", "hola") // This will panic->recover

	mp.unregisterChannel("test2", ch2)
	mp.unregisterChannel("test1", ch1)
	mp.unregisterChannel("test1", ch11)
	require.Equal(t, 0, len(mp.topics))
}

func TestMinipsInt(t *testing.T) {
	mp := NewMinips[int](context.Background())

	ch1 := make(chan int, 100)

	mp.registerChannel("test1", ch1)

	mp.Publish("test1.asdf.qwe", 123)
	require.Equal(t, 1, len(ch1))
}

func TestMinipsUnregAll(t *testing.T) {
	mp := NewMinips[int](context.Background())

	ch1 := make(chan int, 100)

	mp.registerChannel("test1", ch1)
	mp.registerChannel("test2", ch1)
	mp.registerChannel("test3", ch1)
	mp.registerChannel("test4", ch1)

	mp.unregisterChannelFromAll(ch1)

	require.Equal(t, 0, len(mp.topics))
}

func TestSubscriber(t *testing.T) {
	mp := NewMinips[int](context.Background())

	s := mp.NewSubscriber(10, "test1", "test2", "test3")
	mp.Publish("test1", 123)
	mp.Publish("test2", 456)
	mp.Publish("test4", 0)

	require.Equal(t, 2, len(s.Channel()))
	require.Equal(t, 123, <-s.Channel())

	n, err := s.WaitOne(time.Duration(time.Second))
	require.NoError(t, err)
	require.Equal(t, 456, n)

	err = s.Subscribe("test4")
	require.NoError(t, err)
	_, err = s.WaitOne(time.Duration(time.Millisecond * 100))
	require.Error(t, err)

	s.Unsubscribe("test1")
	mp.Publish("test1", 123)
	_, err = s.WaitOne(time.Duration(time.Millisecond * 100))
	require.Error(t, err)

	s.UnsubscribeAll()
	mp.Publish("test2", 123)
	_, err = s.WaitOne(time.Duration(time.Millisecond * 100))
	require.Error(t, err)

	s.Close()
	err = s.Subscribe("test5")
	require.Error(t, err)

	mp.Close()
	ns := mp.NewSubscriber(1, "test")
	_, err = ns.WaitOne(time.Duration(time.Millisecond * 100))
	require.Error(t, err)
}
