package idefixgo

import (
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/garagemakers/idefix/core/cert"
)

func TestPublish(t *testing.T) {
	imc, err := NewClient(context.Background(), &ConnectionOptions{
		BrokerAddress: "tcp://localhost:1883",
		Encoding:      "mg",
		CACert:        cert.CaCert,
		Address:       "jaime-test",
		Token:         "jaime-token",
	})

	if err != nil {
		log.Fatalln(err)
	}

	err = imc.Publish("5c9719505534d914", &Message{
		To:       "asdf",
		Data:     map[string]interface{}{"testing": true},
		Response: "replyhere",
	})

	require.NoError(t, err)

}
