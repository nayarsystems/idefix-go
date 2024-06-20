package main

import (
	"fmt"
	"os"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/spf13/cobra"
)

func init() {
	cmdConfigUpdate.Flags().StringP("file", "f", "", "New configuration file")
	cmdConfigUpdate.MarkFlagRequired("file")

	cmdConfig.AddCommand(cmdConfigUpdate)
}

var cmdConfigUpdate = &cobra.Command{
	Use:   "update",
	Short: "update remote client configuration",
	RunE:  cmdConfigUpdateRunE,
}

func cmdConfigUpdateRunE(cmd *cobra.Command, args []string) (err error) {
	conf := cmdConfigGetBaseParams(cmd)

	// Read the new configuration file
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}
	newConfig, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("cannot read file %s: %w", file, err)
	}

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	msg := &m.AddressConfigUpdateMsg{
		Address: conf.address,
		Config:  newConfig,
	}
	var res m.AddressConfigUpdateResponseMsg
	err = ic.Call2(m.IdefixCmdPrefix, &m.Message{To: m.CmdAddressConfigUpdate, Data: msg}, &res, getTimeout(cmd))
	if err != nil {
		return fmt.Errorf("cannot update client configuration: %w", err)
	}
	fmt.Println("configuration file updated successfully")
	return nil
}
