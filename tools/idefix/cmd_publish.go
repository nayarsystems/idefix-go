package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	idf "gitlab.com/garagemakers/idefix-go"
)

func init() {
	cmdPublish.Flags().StringP("device", "d", "", "Device ID")
	cmdLog.MarkFlagRequired("device")

	rootCmd.AddCommand(cmdPublish)
}

var cmdPublish = &cobra.Command{
	Use:   "publish",
	Short: "Publish a map to the device",
	Args:  cobra.MinimumNArgs(1),
	RunE:  cmdPublishRunE,
}

func cmdPublishRunE(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("device")
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

	if err := ic.Publish(addr, &idf.Message{To: args[0], Data: amap}); err != nil {
		return fmt.Errorf("Cannot publish the message to the device: %w", err)
	}

	return nil
}
