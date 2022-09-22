package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	idf "gitlab.com/garagemakers/idefix-go"
)

func init() {
	cmdCall.Flags().StringP("device", "d", "", "Device ID")
	cmdLog.MarkFlagRequired("device")

	rootCmd.AddCommand(cmdCall)
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

	ret, err := ic.Call(deviceId, &idf.Message{To: topic, Data: amap}, timeout)
	if err != nil {
		spinner.Fail()
		return fmt.Errorf("Cannot publish the message to the device: %w", err)
	}

	if ret.Err != nil {
		spinner.Fail()
		fmt.Println(ret.Err)
	} else {
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

func cmdCallRunE(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("device")
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
