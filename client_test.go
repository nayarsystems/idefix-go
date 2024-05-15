package idefixgo

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	cert "github.com/nayarsystems/cacert-go"
	m "github.com/nayarsystems/idefix-go/messages"
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
		Domain: m.Domain{
			Domain:      testDomain,
			AccessRules: `func check(){ return 2; }`, // 0: not decided, 1: allow, 2: deny
		},
	}
	msg := &m.Message{
		To:   m.CmdDomainCreate,
		Data: domainCreateMsg,
	}
	err := c.Call2("idefix", msg, nil, time.Second)
	fmt.Println("createTestDomain", err)
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
	err := c.Call2("idefix", msg, nil, time.Second)
	fmt.Println("assignTestDomain", err)

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

	_, err = s.WaitOne(time.Second * 5)
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

func TestNormalizeAddress(t *testing.T) {
	tests := []struct {
		address string
		want    string
	}{
		{
			address: "test",
			want:    "test",
		},
		{
			address: "Test123",
			want:    "Test123",
		},
		{
			address: "test@nayarsystems.com",
			want:    "1e2a1b041444207c",
		},
	}

	for _, tt := range tests {
		got := normalizeAddress(tt.address)
		require.Equal(t, tt.want, got)
	}
}
