package main

import (
	"fmt"

	"github.com/spf13/cobra"
	idf "gitlab.com/garagemakers/idefix-go"
)

func init() {
	cmdInfo.Flags().StringP("device", "d", "", "Device ID")
	cmdLog.MarkFlagRequired("device")

	rootCmd.AddCommand(cmdInfo)
}

var cmdInfo = &cobra.Command{
	Use:   "info",
	Short: "Request Device ID info",
	RunE:  cmdInfoRunE,
}

func cmdInfoRunE(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("device")
	if err != nil {
		return err
	}

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	info, err := ic.Call(addr, &idf.Message{To: "sys.cmd.info"}, getTimeout(cmd))
	if err != nil {
		return fmt.Errorf("Cannot get the device info: %w", err)
	}

	if info.Err != nil {
		return fmt.Errorf("Unexpected response: %v", info.Err)
	}

	data, ok := info.Data.(map[string]any)
	if !ok {
		return fmt.Errorf("Unexpected response: %v", info.Data)
	}

	for k, v := range data {
		fmt.Printf("%s: %v\n", k, v)
	}

	return nil
}
