package main

import (
	"encoding/json"
	"fmt"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/spf13/cobra"
)

func init() {
	cmdRulesGet.Flags().StringP("address", "a", "", "Device address")
	cmdRulesGet.MarkFlagRequired("address")
	cmdRules.AddCommand(cmdRulesGet)

	cmdRulesUpdate.Flags().StringP("address", "a", "", "Device address")
	cmdRulesUpdate.MarkFlagRequired("address")
	cmdRulesUpdate.Flags().StringP("allow", "", "", "Allow rules")
	cmdRulesUpdate.Flags().StringP("deny", "", "", "Deny rules")
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
	msg := m.RulesGetMsg{}
	msg.Address, err = cmd.Flags().GetString("address")

	if err != nil {
		return err
	}

	return commandCall2(m.IdefixCmdPrefix, m.CmdAddressRulesGet, msg, getTimeout(cmd))
}

func cmdRulesUpdateRunE(cmd *cobra.Command, args []string) (err error) {
	msg := m.RulesUpdateMsg{}
	msg.Address, err = cmd.Flags().GetString("address")
	if err != nil {
		return err
	}

	sallow, err := cmd.Flags().GetString("allow")
	if err != nil {
		return err
	}
	if cmd.Flags().Changed("allow") {
		dummy := make(map[string]interface{})
		err = json.Unmarshal([]byte(sallow), &dummy)
		if err != nil {
			return fmt.Errorf("cannot parse allow rule: %w", err)
		}
		msg.Allow = sallow
	}

	sdeny, err := cmd.Flags().GetString("deny")
	if err != nil {
		return err
	}
	if cmd.Flags().Changed("deny") {
		dummy := make(map[string]interface{})
		err = json.Unmarshal([]byte(sdeny), &dummy)
		if err != nil {
			return fmt.Errorf("cannot parse deny rule: %w", err)
		}
		msg.Deny = sdeny
	}

	return commandCall2(m.IdefixCmdPrefix, m.CmdAddressRulesUpdate, msg, getTimeout(cmd))
}
