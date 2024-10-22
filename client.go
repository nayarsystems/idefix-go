package idefixgo

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"regexp"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	ie "github.com/nayarsystems/idefix-go/errors"
	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/nayarsystems/idefix-go/minips"
)

// ConnectionStatus represents the current connection state of the Client.
// It is used to indicate whether the Client is connected to the MQTT broker or not.
type ConnectionStatus int64

const (
	Disconnected ConnectionStatus = iota
	Connected
)

// ConnectionStatusHandler is a function type that defines a handler for connection status changes
// in the [Client]. This handler is called whenever the connection status of the [Client] changes,
// allowing users to implement custom behavior based on the new status.
type ConnectionStatusHandler func(*Client, ConnectionStatus)

// Client represents a connection to Idefix, providing methods to interact with. It encapsulates the context, configuration, and connection details necessary for operation.
type Client struct {
	pctx                    context.Context
	ctx                     context.Context
	cancelFunc              context.CancelFunc
	opts                    *ClientOptions
	ps                      *minips.Minips[*m.Message]
	client                  mqtt.Client
	prefix                  string
	sessionID               string
	connectionState         ConnectionStatus
	ConnectionStatusHandler ConnectionStatusHandler
}

// NewClient returns a new [Client] with the options and the context given
func NewClient(pctx context.Context, opts *ClientOptions) *Client {
	c := &Client{
		opts: opts,
		pctx: pctx,
	}

	return c
}

// NewClientFromFile returns a new [Client] with the context given and the options readen from a config file given
func NewClientFromFile(pctx context.Context, configFile string) (*Client, error) {
	opts, err := ReadConfig(configFile)
	if err != nil {
		return nil, err
	}

	c := &Client{
		opts: opts,
		pctx: pctx,
	}

	return c, nil
}

// Connect establishes a connection to the MQTT broker. This method must be called
// when the client is in the Disconnected state. If the client is already connected,
// an error will be returned.
//
// The Connect method initializes the MQTT client with the provided options,
// including the broker address, credentials, and optional TLS configuration if
// a CA certificate is provided. It generates a unique session ID if one is not
// specified in the options.
//
// Upon successful connection, it subscribes to the client's designated
// response topic and performs a login operation. The connection state is then
// updated to Connected.
func (c *Client) Connect() (err error) {
	if c.connectionState != Disconnected {
		return ie.ErrAlreadyExists
	}

	c.ctx, c.cancelFunc = context.WithCancel(c.pctx)

	c.prefix = m.MqttIdefixPrefix
	c.ps = minips.NewMinips[*m.Message](c.ctx)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(c.opts.Broker)
	opts.SetCleanSession(true)
	opts.SetUsername("device")
	opts.SetPassword("77dev22p1")

	if len(c.opts.CACert) > 0 {
		certpool := x509.NewCertPool()
		certpool.AppendCertsFromPEM(c.opts.CACert)

		opts.SetTLSConfig(&tls.Config{
			RootCAs: certpool,
		})
	}

	if c.sessionID == "" {
		if c.opts.SessionID != "" {
			c.sessionID = c.opts.SessionID
		} else {
			c.sessionID, err = randSessionID()
			if err != nil {
				return err
			}
		}
	}
	opts.SetClientID(c.sessionID)

	opts.SetConnectionLostHandler(c.connectionLostHandler)
	opts.SetDefaultPublishHandler(c.receiveMessage)

	c.client = mqtt.NewClient(opts)

	token := c.client.Connect()
	token.Wait()
	if token.Error() != nil {
		return ie.ErrInternal.With(token.Error().Error())
	}

	token = c.client.Subscribe(fmt.Sprintf("%s/%s/r/+", c.prefix, c.sessionID), 1, nil)
	token.Wait()
	if token.Error() != nil {
		return ie.ErrInternal.With(token.Error().Error())
	}

	if err := c.login(); err != nil {
		return err
	}

	c.setState(Connected)
	return nil
}

// SetSessionID sets the sessionID for a client
func (c *Client) SetSessionID(sessionID string) {
	c.sessionID = sessionID
}

// Disconnect gracefully terminates the connection to the MQTT broker.
// This method changes the client's state to Disconnected and invokes
// the cancel function associated with the client's context, which
// may trigger any pending operations or goroutines related to the
// client's connection.
func (c *Client) Disconnect() {
	c.setState(Disconnected)
	c.cancelFunc()
	c.client.Disconnect(200)
}

// Returns the client context
func (c *Client) Context() context.Context {
	return c.ctx
}

// Sets the connection status to the client and updates the [ConnectionStatusHandler] if initialized
func (c *Client) setState(cs ConnectionStatus) {
	if c.connectionState == cs {
		return
	}

	c.connectionState = cs

	if c.ConnectionStatusHandler != nil {
		c.ConnectionStatusHandler(c, c.connectionState)
	}
}

// Returns the connection status of a given client.
func (c *Client) Status() ConnectionStatus {
	return c.connectionState
}

func (c *Client) responseTopic() string {
	return fmt.Sprintf("%s/%s/r/", c.prefix, c.sessionID)
}

func (c *Client) publishTopic(flags string) string {
	return fmt.Sprintf("%s/%s/t/%s", c.prefix, c.sessionID, flags)
}

func (c *Client) login() (err error) {
	lm := m.LoginMsg{
		Address:  c.opts.Address,
		Token:    c.opts.Token,
		Encoding: c.opts.Encoding,
		Meta:     c.opts.Meta,
		Time:     time.Now().UnixMilli(),
		Groups:   c.opts.Groups,
	}

	tm := &m.Message{
		To:   "login",
		Data: lm,
	}

	m, err := c.Call("idefix", tm, time.Second*3)
	if err != nil {
		return err
	}

	if m.Err != "" {
		return fmt.Errorf(m.Err)
	}

	return nil
}

func randSessionID() (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", ie.ErrInternal.Withf("can't create connection ID: %v", err)
	}
	return hex.EncodeToString(b), nil
}

func (c *Client) connectionLostHandler(client mqtt.Client, err error) {
	c.setState(Disconnected)
	c.cancelFunc()
}

// Given an address, this method will return a valid address. If the provided address is valid, the returned address will be the same. Otherwise, if the given address contains characters other than numbers, letters, or dashes, this method will generate the SHA256 hash of that address and return the first 16 characters.
func NormalizeAddress(address string) string {
	r := regexp.MustCompile(`^[a-zA-Z0-9\-]+$`)
	if !r.MatchString(address) {
		hash := sha256.Sum256([]byte(address))
		return hex.EncodeToString(hash[:])[:16]
	}
	return address
}
