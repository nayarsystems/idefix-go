package idefixgo

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nayarsystems/idefix-go/minips"
)

type ConnectionStatus int64

const (
	Disconnected ConnectionStatus = iota
	Connected
)

type ConnectionStatusHandler func(*Client, ConnectionStatus)

type Client struct {
	pctx                    context.Context
	ctx                     context.Context
	cancelFunc              context.CancelFunc
	opts                    *ClientOptions
	ps                      *minips.Minips[*Message]
	client                  mqtt.Client
	prefix                  string
	sessionID               string
	connectionState         ConnectionStatus
	ConnectionStatusHandler ConnectionStatusHandler
}

func NewClient(pctx context.Context, opts *ClientOptions) *Client {
	c := &Client{
		opts: opts,
		pctx: pctx,
	}

	return c
}

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

func (c *Client) Connect() (err error) {
	if c.connectionState != Disconnected {
		return fmt.Errorf("already connected")
	}

	c.ctx, c.cancelFunc = context.WithCancel(c.pctx)

	c.prefix = MqttIdefixPrefix
	c.ps = minips.NewMinips[*Message](c.ctx)

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

	c.sessionID, err = randSessionID()
	if err != nil {
		return err
	}
	opts.SetClientID(c.sessionID)

	opts.SetConnectionLostHandler(c.connectionLostHandler)
	opts.SetDefaultPublishHandler(c.receiveMessage)

	c.client = mqtt.NewClient(opts)

	token := c.client.Connect()
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	token = c.client.Subscribe(fmt.Sprintf("%s/%s/r/+", c.prefix, c.sessionID), 1, nil)
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	if err := c.login(); err != nil {
		return err
	}

	c.setState(Connected)
	return nil
}

func (c *Client) Disconnect() {
	c.setState(Disconnected)
	c.cancelFunc()
	c.client.Disconnect(200)
}

func (c *Client) Context() context.Context {
	return c.ctx
}

func (c *Client) setState(cs ConnectionStatus) {
	if c.connectionState == cs {
		return
	}

	c.connectionState = cs

	if c.ConnectionStatusHandler != nil {
		c.ConnectionStatusHandler(c, c.connectionState)
	}
}

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
	lm := loginMsg{
		Address:  c.opts.Address,
		Token:    c.opts.Token,
		Encoding: c.opts.Encoding,
		Meta:     c.opts.Meta,
		Time:     time.Now().UnixMilli(),
	}

	tm := &Message{
		To:   "login",
		Data: lm,
	}

	m, err := c.Call("idefix", tm, time.Second*3)
	if err != nil {
		return err
	}

	if m.Err != nil {
		return m.Err
	}

	return nil
}

func randSessionID() (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("can't create connection ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (c *Client) connectionLostHandler(client mqtt.Client, err error) {
	c.cancelFunc()
}
