package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/nayarsystems/idefix-go/normalize"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	cmdCall.Flags().StringP("address", "a", "", "Device address")
	cmdCall.MarkFlagRequired("address")

	rootCmd.AddCommand(cmdCall)
	rootCmd.AddCommand(cmdListen)
}

var cmdCall = &cobra.Command{
	Use:   "call",
	Short: "Publish a map to the device and expect an answer",
	Args:  cobra.MinimumNArgs(1),
	RunE:  cmdCallRunE,
}

func commandCall(deviceId string, topic string, amap map[string]interface{}, timeout time.Duration) error {
	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start(fmt.Sprintf("Calling %s@%s with args: %v", topic, deviceId, amap))

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	ret, err := ic.Call(deviceId, &m.Message{To: topic, Data: amap}, timeout)
	if err != nil {
		spinner.Fail()
		return fmt.Errorf("cannot publish the message to the device: %w", err)
	}

	if ret.Err != "" {
		spinner.Fail()
		return fmt.Errorf(ret.Err)
	} else {
		if m, ok := ret.Data.(map[string]interface{}); ok {
			normalize.EncodeTypes(m, &normalize.EncodeTypesOpts{BytesToB64: true})
		}

		rj, err := json.MarshalIndent(ret.Data, "", "  ")
		if err != nil {
			spinner.Fail()
			return err
		}
		spinner.Success()
		fmt.Printf("%s\n", rj)
	}
	return nil
}

func commandCall2(deviceId string, topic string, msg any, timeout time.Duration) error {
	amap, err := m.ToMsi(msg)
	if err != nil {
		return err
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start(fmt.Sprintf("Calling %s@%s with args: %v", topic, deviceId, amap))

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	t0 := time.Now()
	ret, err := ic.Call(deviceId, &m.Message{To: topic, Data: amap}, timeout)
	if err != nil {
		spinner.Fail(fmt.Sprintf("%s (%dms)", spinner.Text, time.Since(t0).Milliseconds()))
		return fmt.Errorf("cannot publish the message to the device: %w", err)
	}

	if ret.Err != "" {
		spinner.Fail(fmt.Sprintf("%s (%dms)", spinner.Text, time.Since(t0).Milliseconds()))
		return fmt.Errorf(ret.Err)
	} else {
		rj, err := json.MarshalIndent(ret.Data, "", "  ")
		if err != nil {
			spinner.Fail(fmt.Sprintf("%s (%dms)", spinner.Text, time.Since(t0).Milliseconds()))
			return err
		}
		spinner.Success(fmt.Sprintf("%s (%dms)", spinner.Text, time.Since(t0).Milliseconds()))
		fmt.Printf("%s\n", rj)
	}
	return nil
}

func cmdCallRunE(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
	}

	amap := make(map[string]interface{})
	if len(args) > 1 {
		if err := json.Unmarshal([]byte(strings.Join(args[1:], " ")), &amap); err != nil {
			return err
		}
	}

	return commandCall(addr, args[0], amap, getTimeout(cmd))
}

var cmdListen = &cobra.Command{
	Use:   "listen",
	Short: "Wait for messages",
	RunE:  cmdListenRunE,
}

func cmdListenRunE(cmd *cobra.Command, args []string) error {
	ic, err := getConnectedClient()
	if err != nil {
		return err
	}

	topic := ""
	if len(args) > 0 {
		topic = args[0]
	}

	sub := ic.NewSubscriber(100, topic)
	defer sub.Close()

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start(fmt.Sprintf(
		"Waiting for messages on topic %q", topic))
	defer spinner.Stop()

	for {
		select {
		case msg := <-sub.Channel():
			fmt.Printf("\nTo: %s, Res: %s, Err: %v, Data: %v\n", msg.To, msg.Res, msg.Err, msg.Data)
		case <-rootctx.Done():
			return nil
		}
	}
}
