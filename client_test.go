package idefixgo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/garagemakers/idefix/core/cert"
)

func TestPublish(t *testing.T) {
	imc, err := NewClient(context.Background(), &ConnectionOptions{
		BrokerAddress: "tcp://localhost:1883",
		Encoding:      "mg",
		CACert:        cert.CaCert,
		Address:       "test",
		Token:         "token",
	})

	require.NoError(t, err)

	s := imc.NewSubscriber(10, "asdf")
	defer s.Close()

	err = imc.Publish("test", &Message{
		To:       "asdf",
		Data:     map[string]interface{}{"testing": true},
		Response: "replyhere",
	})

	require.NoError(t, err)
	time.Sleep(time.Second)
	require.Equal(t, 1, len(s.Channel()))
}

func TestUnauthorized(t *testing.T) {
	_, err := NewClient(context.Background(), &ConnectionOptions{
		BrokerAddress: "tcp://localhost:1883",
		Encoding:      "mg",
		CACert:        cert.CaCert,
		Address:       "test",
		Token:         "tokenn",
	})

	require.Error(t, err)
}

func TestStream(t *testing.T) {
	c, err := NewClient(context.Background(), &ConnectionOptions{
		BrokerAddress: "tcp://localhost:1883",
		Encoding:      "mg",
		CACert:        cert.CaCert,
		Address:       "test",
		Token:         "token",
	})
	require.NoError(t, err)

	c2, err := NewClient(context.Background(), &ConnectionOptions{
		BrokerAddress: "tcp://localhost:1883",
		Encoding:      "mg",
		CACert:        cert.CaCert,
		Address:       "test2",
		Token:         "token",
	})

	require.NoError(t, err)

	s, err := c.NewStream("5c9719505534d914", "asdf", 100, time.Minute)
	require.NoError(t, err)
	defer s.Close()

	err = c2.Publish("5c9719505534d914", &Message{To: "asdf", Data: "test"})
	require.NoError(t, err)

	e := <-s.Channel()
	require.Equal(t, "test", e.Data)

}
