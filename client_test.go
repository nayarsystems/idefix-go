package idefixgo

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/nayarsystems/idefix/core/cert"
	"github.com/stretchr/testify/require"
)

const testDomain string = "test"

func TestMain(m *testing.M) {

	// Run tests
	err := setup()
	code := 0
	if err == nil {
		code = m.Run()
	} else {
		fmt.Printf("can't perform client tests: %v\n", err)
	}

	os.Exit(code)
}

func setup() error {
	c1 := createTestClient("test1", "test1token")
	err := c1.Connect()
	if err != nil {
		return fmt.Errorf("backend not available: %v", err)
	}
	defer c1.Disconnect()
	createTestDomain(c1)
	assignTestDomain(c1, c1)
	return nil
}

func createTestDomain(c *Client) {
	domainCreateMsg := m.DomainCreateMsg{
		DomainInfo: m.DomainInfo{
			Domain: testDomain,
			Allow:  `{"$true": true}`,
		},
	}
	msg := &m.Message{
		To:   m.CmdDomainCreate,
		Data: domainCreateMsg,
	}
	_ = c.Call2("idefix", msg, nil, time.Second)
}

func assignTestDomain(c *Client, d *Client) {
	domainAssignMsg := m.DomainAssignMsg{
		Address: d.opts.Address,
		Domain:  testDomain,
	}

	msg := &m.Message{
		To:   m.CmdDomainAssign,
		Data: domainAssignMsg,
	}
	_ = c.Call2("idefix", msg, nil, time.Second)
}

func createTestClient(address, token string) *Client {
	testClient := NewClient(context.Background(), &ClientOptions{
		Broker:   "tcp://localhost:1883",
		Encoding: "mg",
		CACert:   cert.CaCert,
		Address:  address,
		Token:    token,
	})
	return testClient
}
func TestPublish(t *testing.T) {
	c1 := createTestClient("test1", "test1token")
	c1.Connect()
	defer c1.Disconnect()

	s := c1.NewSubscriber(10, "asdf")
	defer s.Close()

	err := c1.Publish(c1.opts.Address, &m.Message{
		To:   "asdf",
		Data: map[string]interface{}{"testing": true},
		Res:  "replyhere",
	})

	require.NoError(t, err)

	_, err = s.WaitOne(time.Second)
	require.NoError(t, err)
}

func TestUnauthorized(t *testing.T) {
	c1 := createTestClient("unauthorized", "unauthorizedToken")
	err := c1.Connect()
	require.NoError(t, err)
	c1.Disconnect()

	c2 := createTestClient("unauthorized", "unauthorizedWrongToken")
	err = c2.Connect()
	require.Error(t, err)
}

func TestConnectionHandler(t *testing.T) {
	c := createTestClient("test", "testToken")

	statuses := []ConnectionStatus{}

	c.ConnectionStatusHandler = func(c *Client, cs ConnectionStatus) {
		statuses = append(statuses, cs)
	}

	err := c.Connect()
	require.NoError(t, err)
	require.Equal(t, Connected, c.Status())

	err = c.Connect()
	require.Error(t, err)

	c.Disconnect()

	err = c.Connect()
	require.NoError(t, err)

	require.Equal(t, []ConnectionStatus{Connected, Disconnected, Connected}, statuses)
}
