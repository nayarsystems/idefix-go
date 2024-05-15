package main

import (
	idf "github.com/nayarsystems/idefix-go"
	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/spf13/cobra"
)

func init() {
	cmdGroupsGet.Flags().StringP("address", "a", "", "Device address")
	_ = cmdGroupsGet.MarkFlagRequired("address")
	cmdGroupsGet.Flags().StringP("domain", "d", "", "Domain")
	cmdGroups.AddCommand(cmdGroupsGet)

	cmdGroupsAdd.Flags().StringP("address", "a", "", "Device address")
	_ = cmdGroupsAdd.MarkFlagRequired("address")
	cmdGroupsAdd.Flags().StringP("domain", "d", "", "Domain")
	_ = cmdGroupsAdd.MarkFlagRequired("domain")
	cmdGroupsAdd.Flags().StringP("group", "g", "", "Group")
	_ = cmdGroupsAdd.MarkFlagRequired("group")
	cmdGroups.AddCommand(cmdGroupsAdd)

	cmdGroupsRemove.Flags().StringP("address", "a", "", "Device address")
	_ = cmdGroupsRemove.MarkFlagRequired("address")
	cmdGroupsRemove.Flags().StringP("domain", "d", "", "Domain")
	_ = cmdGroupsRemove.MarkFlagRequired("domain")
	cmdGroupsRemove.Flags().StringP("group", "g", "", "Group")
	_ = cmdGroupsRemove.MarkFlagRequired("group")
	cmdGroups.AddCommand(cmdGroupsRemove)

	rootCmd.AddCommand(cmdGroups)
}

var cmdGroups = &cobra.Command{
	Use:   "groups",
	Short: "Client groups management",
}

var cmdGroupsGet = &cobra.Command{
	Use:   "get",
	Short: "Show groups of an address at a domain",
	RunE:  cmdGroupsGetRunE,
}

var cmdGroupsAdd = &cobra.Command{
	Use:   "add",
	Short: "Add an address to a group",
	RunE:  cmdGroupsAddRunE,
}

var cmdGroupsRemove = &cobra.Command{
	Use:   "remove",
	Short: "Remove an address from a group",
	RunE:  cmdGroupsRemoveRunE,
}

func cmdGroupsGetRunE(cmd *cobra.Command, args []string) (err error) {
	msg := m.AddressGetGroupsMsg{}

	msg.Address, err = cmd.Flags().GetString("address")
	if err != nil {
		return err
	}
	msg.Address = idf.NormalizeAddress(msg.Address)

	msg.Domain, _ = cmd.Flags().GetString("domain")

	return commandCall2(m.IdefixCmdPrefix, m.CmdAddressGetGroups, msg, getTimeout(cmd))
}

func cmdGroupsAddRunE(cmd *cobra.Command, args []string) (err error) {
	msg := m.GroupAddAddressMsg{}
	msg.Address, err = cmd.Flags().GetString("address")
	if err != nil {
		return err
	}
	msg.Address = idf.NormalizeAddress(msg.Address)

	msg.Domain, err = cmd.Flags().GetString("domain")
	if err != nil {
		return err
	}

	msg.Group, err = cmd.Flags().GetString("group")
	if err != nil {
		return err
	}

	return commandCall2(m.IdefixCmdPrefix, m.CmdGroupAddAddress, msg, getTimeout(cmd))
}

func cmdGroupsRemoveRunE(cmd *cobra.Command, args []string) (err error) {
	msg := m.GroupRemoveAddressMsg{}
	msg.Address, err = cmd.Flags().GetString("address")
	if err != nil {
		return err
	}
	msg.Address = idf.NormalizeAddress(msg.Address)

	msg.Domain, err = cmd.Flags().GetString("domain")
	if err != nil {
		return err
	}

	msg.Group, err = cmd.Flags().GetString("group")
	if err != nil {
		return err
	}

	return commandCall2(m.IdefixCmdPrefix, m.CmdGroupRemoveAddress, msg, getTimeout(cmd))
}
