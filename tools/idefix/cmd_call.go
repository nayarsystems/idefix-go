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
	cmdCall.Flags().StringP("timeout", "t", "", "Timeout in ms")

	rootCmd.AddCommand(cmdCall)
}

var cmdCall = &cobra.Command{
	Use:   "call",
	Short: "Publish a map to the device and expect an answer",
	Args:  cobra.MinimumNArgs(1),
	RunE:  cmdCallRunE,
}

func cmdCallRunE(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("device")
	if err != nil {
		return err
	}

	timeout := time.Second * 5
	ptimeout, err := cmd.Flags().GetUint("timeout")
	if err == nil {
		timeout = time.Duration(ptimeout * 1000000)
	}

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	amap := make(map[string]interface{})
	if len(args) > 1 {
		if err := json.Unmarshal([]byte(strings.Join(args[1:], " ")), &amap); err != nil {
			return err
		}
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start("Calling ", args[0])
	ret, err := ic.Call(addr, &idf.Message{To: args[0], Data: amap}, timeout)
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
		fmt.Printf("%s", rj)
	}

	return nil
}
