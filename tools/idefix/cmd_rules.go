package main

import (
	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/spf13/cobra"
)

func init() {
	cmdRulesGet.Flags().StringP("address", "a", "", "Device address")
	cmdRulesGet.MarkFlagRequired("address")
	cmdRules.AddCommand(cmdRulesGet)

	cmdRulesUpdate.Flags().StringP("address", "a", "", "Device address")
	cmdRulesUpdate.MarkFlagRequired("address")
	cmdRulesUpdate.Flags().StringP("accessRules", "", "", "Device access rules (snippet)")
	cmdRules.AddCommand(cmdRulesUpdate)

	rootCmd.AddCommand(cmdRules)
}

var cmdRules = &cobra.Command{
	Use:   "rules",
	Short: "Client access rules management",
}

var cmdRulesGet = &cobra.Command{
	Use:   "get",
	Short: "Show access rules management",
	RunE:  cmdRulesGetRunE,
}

var cmdRulesUpdate = &cobra.Command{
	Use:   "update",
	Short: "Update access rules management",
	RunE:  cmdRulesUpdateRunE,
}

func cmdRulesGetRunE(cmd *cobra.Command, args []string) (err error) {
	msg := m.AddressAccessRulesGetMsg{}
	msg.Address, err = cmd.Flags().GetString("address")

	if err != nil {
		return err
	}

	return commandCall2(m.IdefixCmdPrefix, m.CmdAddressRulesGet, msg, getTimeout(cmd))
}

func cmdRulesUpdateRunE(cmd *cobra.Command, args []string) (err error) {
	msg := m.AddressAccessRulesUpdateMsg{}
	msg.Address, err = cmd.Flags().GetString("address")
	if err != nil {
		return err
	}

	snippet, err := cmd.Flags().GetString("accessRules")
	if err != nil {
		return err
	}
	msg.AccessRules = snippet

	return commandCall2(m.IdefixCmdPrefix, m.CmdAddressRulesUpdate, msg, getTimeout(cmd))
}
