package main

import (
	"encoding/json"
	"fmt"
	"strings"

	idf "github.com/nayarsystems/idefix-go"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	cmdPublish.Flags().StringP("address", "a", "", "Device address")
	cmdPublish.MarkFlagRequired("address")

	rootCmd.AddCommand(cmdPublish)
}

var cmdPublish = &cobra.Command{
	Use:   "publish",
	Short: "Publish a map to the device",
	Args:  cobra.MinimumNArgs(1),
	RunE:  cmdPublishRunE,
}

func cmdPublishRunE(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
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

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start("Publishing on ", args[0])
	if err := ic.Publish(addr, &idf.Message{To: args[0], Data: amap}); err != nil {
		spinner.Fail()
		return fmt.Errorf("Cannot publish the message to the device: %w", err)
	}
	spinner.Success()

	return nil
}
