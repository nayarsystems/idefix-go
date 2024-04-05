package main

import (
	"encoding/json"
	"fmt"
	"strings"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	cmdPublish.Flags().StringP("address", "a", "", "Device address")
	cmdPublish.Flags().BoolP("bool", "b", false, "Is boolean")
	cmdPublish.MarkFlagRequired("address")

	rootCmd.AddCommand(cmdPublish)
}

var cmdPublish = &cobra.Command{
	Use:   "publish",
	Short: "Publish a value to the device",
	Args:  cobra.MinimumNArgs(1),
	RunE:  cmdPublishRunE,
}

func cmdPublishRunE(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
	}
	isBool, err := cmd.Flags().GetBool("bool")
	if err != nil {
		return err
	}

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	var value any
	if len(args) > 1 {
		if isBool {
			if args[1] == "true" || args[1] == "True" || args[1] == "TRUE" || args[1] == "1" {
				value = true
			} else if args[1] == "false" || args[1] == "False" || args[1] == "FALSE" || args[1] == "0" {
				value = false
			} else {
				return fmt.Errorf("invalid boolean value")
			}
		} else {
			amap := make(map[string]interface{})
			if err := json.Unmarshal([]byte(strings.Join(args[1:], " ")), &amap); err != nil {
				if len(args) > 2 {
					return fmt.Errorf("invalid arguments")
				}
				value = args[1]
			} else {
				// Consider a string value
				value = amap
			}
		}
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start("Publishing on ", args[0])
	if err := ic.Publish(addr, &m.Message{To: args[0], Data: value}); err != nil {
		spinner.Fail()
		return fmt.Errorf("cannot publish the message to the device: %w", err)
	}
	spinner.Success()

	return nil
}
