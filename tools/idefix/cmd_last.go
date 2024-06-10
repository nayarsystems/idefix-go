package main

import (
	"encoding/json"
	"fmt"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	cmdLast.Flags().StringP("address", "a", "", "Device address")
	cmdLast.MarkFlagRequired("address")

	rootCmd.AddCommand(cmdLast)
}

var cmdLast = &cobra.Command{
	Use:   "last",
	Short: "Request last state published by the device",
	RunE:  cmdLastRunE,
}

func cmdLastRunE(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
	}

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	msg := &m.AddressStatesGetMsg{
		Address: addr,
	}
	var res m.AddressStatesGetResMsg
	err = ic.Call2(m.IdefixCmdPrefix, &m.Message{To: m.CmdAddressStatesGet, Data: msg}, &res, getTimeout(cmd))
	if err != nil {
		return fmt.Errorf("cannot get the device last state: %w", err)
	}

	if len(res.States) > 0 {
		if b, err := json.MarshalIndent(res.States, "", "  "); err == nil {
			pterm.Info.Printf("Last states:\n%s\n", b)
		}
	}

	fmt.Println()
	return nil
}
