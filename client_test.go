package idefixgo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/garagemakers/idefix/core/cert"
)

func TestPublish(t *testing.T) {
	imc := NewClient(context.Background(), &ClientOptions{
		BrokerAddress: "tcp://localhost:1883",
		Encoding:      "mg",
		CACert:        cert.CaCert,
		Address:       "test",
		Token:         "token",
	})
	err := imc.Connect()
	require.NoError(t, err)

	s := imc.NewSubscriber(10, "asdf")
	defer s.Close()

	err = imc.Publish("test", &Message{
		To:       "asdf",
		Data:     map[string]interface{}{"testing": true},
		Response: "replyhere",
	})

	require.NoError(t, err)

	_, err = s.WaitOne(time.Second)
	require.NoError(t, err)
}

func TestUnauthorized(t *testing.T) {
	c := NewClient(context.Background(), &ClientOptions{
		BrokerAddress: "tcp://localhost:1883",
		Encoding:      "mg",
		CACert:        cert.CaCert,
		Address:       "test11",
		Token:         "tokenn12312n",
	})

	err := c.Connect()
	require.Error(t, err)
}

func TestStream(t *testing.T) {
	c := NewClient(context.Background(), &ClientOptions{
		BrokerAddress: "tcp://localhost:1883",
		Encoding:      "mg",
		CACert:        cert.CaCert,
		Address:       "test",
		Token:         "token",
	})

	err := c.Connect()
	require.NoError(t, err)

	c2 := NewClient(context.Background(), &ClientOptions{
		BrokerAddress: "tcp://localhost:1883",
		Encoding:      "mg",
		CACert:        cert.CaCert,
		Address:       "test2",
		Token:         "token",
	})

	err = c2.Connect()
	require.NoError(t, err)

	s, err := c.NewStream("5c9719505534d914", "asdf", 100, time.Minute)
	require.NoError(t, err)
	defer s.Close()

	err = c2.Publish("5c9719505534d914", &Message{To: "asdf", Data: "test"})
	require.NoError(t, err)

	e := <-s.Channel()
	require.Equal(t, "test", e.Data)

}
