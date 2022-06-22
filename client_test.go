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
