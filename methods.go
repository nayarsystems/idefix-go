package idefixgo

import (
	"context"
	"fmt"
	"time"

	"github.com/jaracil/ei"
	ie "github.com/nayarsystems/idefix-go/errors"
	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/nayarsystems/idefix-go/minips"
)

// Publish sends a message to a specified remote address.
func (c *Client) Publish(remoteAddress string, msg *m.Message) error {
	msg.To = fmt.Sprintf("%s.%s", remoteAddress, msg.To)
	return c.sendMessage(msg)
}

// Answer constructs a response message based on the original message
// and sends it back to the intended recipient.
//
// The original message, 'origmsg', is used to determine the response
// destination by setting the 'To' field of the response message `msg`
// to the 'Res' field of the original message. This allows the sender
// of the original message to receive the response correctly.
//
// It is important to ensure that 'origmsg' contains a valid response
// destination in its 'Res' field before calling this method. If the
// response message fails to send, an error will be returned.
func (c *Client) Answer(origmsg *m.Message, msg *m.Message) error {
	msg.To = origmsg.Res // TODO: Check this
	return c.sendMessage(msg)
}

// Call sends a message to a specified remote address and expects a response. If timeout given is exceed it returns an error.
func (c *Client) Call(remoteAddress string, msg *m.Message, timeout time.Duration) (*m.Message, error) {
	var err error
	msg.To = fmt.Sprintf("%s.%s", remoteAddress, msg.To)
	msg.Res, err = randSessionID()
	if err != nil {
		return nil, err
	}

	sub := c.ps.NewSubscriber(1, msg.Res)
	defer sub.Close()

	if err := c.sendMessage(msg); err != nil {
		return nil, err
	}
	msg, err = sub.WaitOne(timeout)
	if err != nil {
		return nil, ie.ErrTimeout
	}
	if msg.Err != "" {
		return msg, fmt.Errorf("%s", msg.Err)
	}
	return msg, nil
}

// CallWithContext sends a message to a specified remote address and expects a response until the context gets cancelled
func (c *Client) CallWithContext(ctx context.Context, remoteAddress string, msg *m.Message) (*m.Message, error) {
	var err error
	msg.To = fmt.Sprintf("%s.%s", remoteAddress, msg.To)
	msg.Res, err = randSessionID()
	if err != nil {
		return nil, err
	}

	sub := c.ps.NewSubscriber(1, msg.Res)
	defer sub.Close()

	if err := c.sendMessageWithContext(ctx, msg); err != nil {
		return nil, err
	}
	msg, err = sub.WaitOneWithContext(ctx)
	if err != nil {
		return nil, ie.ErrTimeout
	}
	if msg.Err != "" {
		return msg, fmt.Errorf("%s", msg.Err)
	}
	return msg, nil
}

// Call2 uses [Client.Call] to send a message to a specified remote address and expects a response.
// The function converts the 'msg.Data' field into a map (msi) format before sending it using the internal [Client.Call] method.
// If a response is expected ('resp' is not nil), it parses the returned data into the provided response structure.
// The function also handles any errors that occur during the process, either in the message sending or the response.
func (c *Client) Call2(remoteAddress string, msg *m.Message, resp any, timeout time.Duration) error {
	amap, err := m.ToMsi(msg.Data)
	if err != nil {
		return err
	}
	msg.Data = amap
	ret, err := c.Call(remoteAddress, msg, timeout)
	if err != nil {
		return err
	}
	if ret.Err != "" {
		return fmt.Errorf("%s", msg.Err)
	}
	if resp != nil {
		respMsi, err := m.GetMsi(ret.Data)
		if err != nil {
			return err
		}
		return m.ParseMsi(respMsi, resp)
	}
	return nil
}

// NewSubscriber creates a new message subscriber with the specified capacity and topic(s).
//
// The function initializes a new subscriber using the client's internal publish-subscribe system ('ps'),
// allowing it to receive messages published on the specified topics. The subscriber is configured with a given
// buffer capacity to manage the number of messages it can hold before processing.
func (c *Client) NewSubscriber(capacity uint, topic ...string) *minips.Subscriber[*m.Message] {
	return c.ps.NewSubscriber(capacity, topic...)
}

// WaitOne waits for a single message on the specified topic within the given timeout duration.
//
// The function subscribes to the specified topic and blocks until a message is received or the timeout occurs.
// If a message is received within the timeout, it is returned. If the timeout expires before a message arrives,
// an error is returned.
func (c *Client) WaitOne(topic string, timeout time.Duration) (*m.Message, error) {
	sub := c.ps.NewSubscriber(1, topic)
	defer sub.Close()
	return sub.WaitOne(timeout)
}

// WaitOneWithContext waits for a single message on the specified topic until the context gets cancelled.
//
// The function subscribes to the specified topic and blocks until a message is received or the context is cancelled.
// If a message is received before the context is cancelled, it is returned. If the context is cancelled before a message
// arrives, an error is returned.
func (c *Client) WaitOneWithContext(ctx context.Context, topic string) (*m.Message, error) {
	sub := c.ps.NewSubscriber(1, topic)
	defer sub.Close()
	return sub.WaitOneWithContext(ctx)
}

func (c *Client) Syscall(message *m.Message, response any, ctx ...context.Context) (err error) {
	var callCtx context.Context
	if ctx == nil || len(ctx) == 0 {
		var cancel context.CancelFunc
		callCtx, cancel = context.WithDeadline(c.ctx, time.Now().Add(time.Second*10))
		defer cancel()
	} else {
		callCtx = ctx[0]
	}

	message.Data, err = m.ToMsi(message.Data)
	if err != nil {
		return err
	}

	ret, err := c.CallWithContext(callCtx, "idefix", message)
	if err != nil {
		return err
	}

	if ret.Err != "" {
		return fmt.Errorf("%s", ret.Err)
	}

	if response != nil {
		if rrr, err := ei.N(ret.Data).Bool(); err == nil {
			response = rrr
			return nil
		}

		respMsi, err := m.GetMsi(ret.Data)
		if err != nil {
			return err
		}
		return m.ParseMsi(respMsi, response)
	}

	return nil
}

func (c *Client) EventCreate(query *m.EventMsg, ctx ...context.Context) (response *m.EventResponseMsg, err error) {
	message := &m.Message{To: m.CmdEventsCreate, Data: query}
	response = &m.EventResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) EventsGet(query *m.EventsGetMsg, ctx ...context.Context) (response *m.EventsGetResponseMsg, err error) {
	message := &m.Message{To: m.CmdEventsGet, Data: query}
	response = &m.EventsGetResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) AddressTokenReset(query *m.AddressTokenResetMsg, ctx ...context.Context) (response bool, err error) {
	message := &m.Message{To: m.CmdAddressTokenReset, Data: query}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) AddressDisable(query *m.AddressDisableMsg, ctx ...context.Context) (response bool, err error) {
	message := &m.Message{To: m.CmdAddressDisable, Data: query}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) AddressAccessRulesGet(query *m.AddressAccessRulesGetMsg, ctx ...context.Context) (response *m.AddressAccessRulesGetResponseMsg, err error) {
	message := &m.Message{To: m.CmdAddressRulesGet, Data: query}
	response = &m.AddressAccessRulesGetResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) AddressAccessRulesUpdate(query *m.AddressAccessRulesUpdateMsg, ctx ...context.Context) (response *m.AddressAccessRulesUpdateResponseMsg, err error) {
	message := &m.Message{To: m.CmdDomainUpdateAccessRules, Data: query}
	response = &m.AddressAccessRulesUpdateResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) AddressDomainGet(query *m.AddressDomainGetMsg, ctx ...context.Context) (response *m.Domain, err error) {
	message := &m.Message{To: m.CmdDomainGet, Data: query}
	response = &m.Domain{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) AddressConfigGet(query *m.AddressConfigGetMsg, ctx ...context.Context) (response *m.AddressConfigGetResponseMsg, err error) {
	message := &m.Message{To: m.CmdAddressConfigGet, Data: query}
	response = &m.AddressConfigGetResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}
func (c *Client) AddressStatesGet(query *m.AddressStatesGetMsg, ctx ...context.Context) (response *m.AddressStatesGetResMsg, err error) {
	message := &m.Message{To: m.CmdAddressConfigGet, Data: query}
	response = &m.AddressStatesGetResMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) AddressConfigUpdate(query *m.AddressConfigUpdateMsg, ctx ...context.Context) (response *m.AddressConfigUpdateResponseMsg, err error) {
	message := &m.Message{To: m.CmdAddressConfigUpdate, Data: query}
	response = &m.AddressConfigUpdateResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) AddressAliasGet(query *m.AddressAliasGetMsg, ctx ...context.Context) (response *m.AddressAliasGetResponseMsg, err error) {
	message := &m.Message{To: m.CmdAddressAliasGet, Data: query}
	response = &m.AddressAliasGetResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) AddressAliasAdd(query *m.AddressAliasAddMsg, ctx ...context.Context) (response *m.AddressAliasAddResponseMsg, err error) {
	message := &m.Message{To: m.CmdAddressAliasAdd, Data: query}
	response = &m.AddressAliasAddResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) AddressAliasRemove(query *m.AddressAliasRemoveMsg, ctx ...context.Context) (response *m.AddressAliasRemoveResponseMsg, err error) {
	message := &m.Message{To: m.CmdAddressAliasRemove, Data: query}
	response = &m.AddressAliasRemoveResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) SchemaCreate(query *m.SchemaMsg, ctx ...context.Context) (response *m.SchemaResponseMsg, err error) {
	message := &m.Message{To: m.CmdSchemasGet, Data: query}
	response = &m.SchemaResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) SchemaGet(query *m.SchemaGetMsg, ctx ...context.Context) (response *m.SchemaGetResponseMsg, err error) {
	message := &m.Message{To: m.CmdSchemasGet, Data: query}
	response = &m.SchemaGetResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) DomainGet(query *m.DomainGetMsg, ctx ...context.Context) (response *m.Domain, err error) {
	message := &m.Message{To: m.CmdDomainGet, Data: query}
	response = &m.Domain{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) DomainDelete(query *m.DomainDeleteMsg, ctx ...context.Context) (response bool, err error) {
	message := &m.Message{To: m.CmdDomainDelete, Data: query}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) DomainCreate(query *m.DomainCreateMsg, ctx ...context.Context) (response *m.DomainCreateResponseMsg, err error) {
	message := &m.Message{To: m.CmdDomainCreate, Data: query}
	response = &m.DomainCreateResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) DomainUpdate(query *m.DomainUpdateMsg, ctx ...context.Context) (response *m.DomainUpdateResponseMsg, err error) {
	message := &m.Message{To: m.CmdDomainUpdate, Data: query}
	response = &m.DomainUpdateResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) DomainUpdateAccessRules(query *m.DomainUpdateAccessRulesMsg, ctx ...context.Context) (response *m.DomainUpdateAccessRulesResponseMsg, err error) {
	message := &m.Message{To: m.CmdDomainUpdateAccessRules, Data: query}
	response = &m.DomainUpdateAccessRulesResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) DomainAssign(query *m.DomainAssignMsg, ctx ...context.Context) (response bool, err error) {
	message := &m.Message{To: m.CmdDomainAssign, Data: query}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) DomainGetTree(query *m.DomainGetTreeMsg, ctx ...context.Context) (response []string, err error) {
	message := &m.Message{To: m.CmdDomainTree, Data: query}
	response = make([]string, 0)
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) DomainCountAddresses(query *m.DomainCountAddressesMsg, ctx ...context.Context) (response *m.DomainCountAddressesResponseMsg, err error) {
	message := &m.Message{To: m.CmdDomainCountAddresses, Data: query}
	response = &m.DomainCountAddressesResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) DomainListAddresses(query *m.DomainListAddressesMsg, ctx ...context.Context) (response *m.DomainListAddressesResponseMsg, err error) {
	message := &m.Message{To: m.CmdDomainListAddresses, Data: query}
	response = &m.DomainListAddressesResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) GroupAddAddress(query *m.GroupAddAddressMsg, ctx ...context.Context) (response *m.GroupAddAddressResponseMsg, err error) {
	message := &m.Message{To: m.CmdGroupAddAddress, Data: query}
	response = &m.GroupAddAddressResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) GroupRemoveAddress(query *m.GroupRemoveAddressMsg, ctx ...context.Context) (response *m.GroupRemoveAddressResponseMsg, err error) {
	message := &m.Message{To: m.CmdGroupRemoveAddress, Data: query}
	response = &m.GroupRemoveAddressResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) GroupGetAddresses(query *m.GroupGetAddressesMsg, ctx ...context.Context) (response *m.GroupGetAddressesResponseMsg, err error) {
	message := &m.Message{To: m.CmdGroupGetAddresses, Data: query}
	response = &m.GroupGetAddressesResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) DomainGetGroups(query *m.DomainGetGroupsMsg, ctx ...context.Context) (response *m.DomainGetGroupsResponseMsg, err error) {
	message := &m.Message{To: m.CmdDomainListGroups, Data: query}
	response = &m.DomainGetGroupsResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) AddressGetGroups(query *m.AddressGetGroupsMsg, ctx ...context.Context) (response *m.AddressGetGroupsResponseMsg, err error) {
	message := &m.Message{To: m.CmdAddressGetGroups, Data: query}
	response = &m.AddressGetGroupsResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) GroupRemove(query *m.GroupRemoveMsg, ctx ...context.Context) (response *m.GroupRemoveResponseMsg, err error) {
	message := &m.Message{To: m.CmdGroupRemoveAddress, Data: query}
	response = &m.GroupRemoveResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) SessionDelete(query *m.SessionDeleteMsg, ctx ...context.Context) (response *m.SessionDeleteResponseMsg, err error) {
	message := &m.Message{To: m.CmdSessionDelete, Data: query}
	response = &m.SessionDeleteResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) AddressEnvironmentGet(query *m.AddressEnvironmentGetMsg, ctx ...context.Context) (response *m.AddressEnvironmentGetResponseMsg, err error) {
	message := &m.Message{To: m.CmdAddressEnvironmentGet, Data: query}
	response = &m.AddressEnvironmentGetResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) AddressEnvironmentSet(query *m.AddressEnvironmentSetMsg, ctx ...context.Context) (response *m.AddressEnvironmentSetResponseMsg, err error) {
	message := &m.Message{To: m.CmdAddressEnvironmentSet, Data: query}
	response = &m.AddressEnvironmentSetResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) AddressEnvironmentUnset(query *m.AddressEnvironmentUnsetMsg, ctx ...context.Context) (response *m.AddressEnvironmentUnsetResponseMsg, err error) {
	message := &m.Message{To: m.CmdAddressEnvironmentUnset, Data: query}
	response = &m.AddressEnvironmentUnsetResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) DomainEnvironmentGet(query *m.DomainEnvironmentGetMsg, ctx ...context.Context) (response *m.DomainEnvironmentGetResponseMsg, err error) {
	message := &m.Message{To: m.CmdDomainEnvironmentGet, Data: query}
	response = &m.DomainEnvironmentGetResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) DomainEnvironmentSet(query *m.DomainEnvironmentSetMsg, ctx ...context.Context) (response *m.DomainEnvironmentSetResponseMsg, err error) {
	message := &m.Message{To: m.CmdDomainEnvironmentSet, Data: query}
	response = &m.DomainEnvironmentSetResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) DomainEnvironmentUnset(query *m.DomainEnvironmentUnsetMsg, ctx ...context.Context) (response *m.DomainEnvironmentUnsetResponseMsg, err error) {
	message := &m.Message{To: m.CmdDomainEnvironmentUnset, Data: query}
	response = &m.DomainEnvironmentUnsetResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}

func (c *Client) Environment(query *m.EnvironmentGetMsg, ctx ...context.Context) (response *m.EnvironmentGetResponseMsg, err error) {
	message := &m.Message{To: m.CmdEnvironmentGet, Data: query}
	response = &m.EnvironmentGetResponseMsg{}
	err = c.Syscall(message, response, ctx...)
	return
}
