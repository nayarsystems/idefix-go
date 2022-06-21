package minips

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMinips(t *testing.T) {
	mp := NewMinips[string]()

	ch1 := make(chan string, 100)
	ch11 := make(chan string, 100)
	ch2 := make(chan string, 100)

	mp.RegisterChannel("test1", ch1)
	mp.RegisterChannel("test2", ch2)
	mp.RegisterChannel("test1", ch11)

	mp.Publish("test1.asdf.qwe", "hola")
	mp.Publish("test1.asdf", "hola")
	mp.Publish("test1", "hola")
	mp.Publish("test2", "hola")

	require.Equal(t, 3, len(ch1))
	require.Equal(t, 3, len(ch11))
	require.Equal(t, 1, len(ch2))

	mp.UnregisterChannel("test2", ch2)
	mp.Publish("test2", "hola")
	require.Equal(t, 1, len(ch2))

	close(ch2)
	mp.RegisterChannel("test2", ch2)
	mp.Publish("test2", "hola") // This will panic->recover

	mp.UnregisterChannel("test2", ch2)
	mp.UnregisterChannel("test1", ch1)
	mp.UnregisterChannel("test1", ch11)
	require.Equal(t, 0, len(mp.topics))
}

func TestMinipsInt(t *testing.T) {
	mp := NewMinips[int]()

	ch1 := make(chan int, 100)

	mp.RegisterChannel("test1", ch1)

	mp.Publish("test1.asdf.qwe", 123)
	require.Equal(t, 1, len(ch1))
}
