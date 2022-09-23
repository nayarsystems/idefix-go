package main

import (
	"encoding/json"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	cmdDomainCreate.Flags().StringP("name", "n", "", "Domain Name")
	cmdDomainCreate.Flags().StringP("allow", "a", "", "Allow rule")
	cmdDomainCreate.Flags().StringP("deny", "d", "", "Deny rule")
	cmdDomainCreate.Flags().StringP("env", "e", "", "Environment Map")
	cmdDomainCreate.Flags().StringArrayP("admin", "", []string{}, "Admin Address (Can be set multiple times)")
	cmdDomainCreate.MarkFlagRequired("name")
	cmdDomain.AddCommand(cmdDomainCreate)

	cmdDomainAssign.Flags().StringP("name", "n", "", "Domain Name")
	cmdDomainAssign.Flags().StringP("address", "a", "", "Device address")
	cmdDomainAssign.MarkFlagRequired("name")
	cmdDomainAssign.MarkFlagRequired("address")
	cmdDomain.AddCommand(cmdDomainAssign)

	cmdDomainGet.Flags().StringP("name", "n", "", "Domain Name")
	cmdDomainGet.MarkFlagRequired("name")
	cmdDomain.AddCommand(cmdDomainGet)

	cmdDomainDelete.Flags().StringP("name", "n", "", "Domain Name")
	cmdDomainDelete.MarkFlagRequired("name")
	cmdDomain.AddCommand(cmdDomainDelete)

	cmdDomainUpdate.Flags().StringP("name", "n", "", "Domain Name")
	cmdDomainUpdate.Flags().StringP("allow", "a", "", "Allow rule")
	cmdDomainUpdate.Flags().StringP("deny", "d", "", "Deny rule")
	cmdDomainUpdate.Flags().StringP("env", "e", "", "Environment Map")
	cmdDomainUpdate.Flags().StringArrayP("admin", "", []string{}, "Admin Address (Can be set multiple times)")
	cmdDomainUpdate.MarkFlagRequired("name")
	cmdDomain.AddCommand(cmdDomainUpdate)

	rootCmd.AddCommand(cmdDomain)
}

var cmdDomain = &cobra.Command{
	Use:     "domain",
	Aliases: []string{"domains"},
	Short:   "Manage idefix domains",
}

var cmdDomainGet = &cobra.Command{
	Use:   "get",
	Short: "Get a domain",
	RunE:  cmdDomainGetRunE,
}

var cmdDomainDelete = &cobra.Command{
	Use:   "delete",
	Short: "Delete a domain",
	RunE:  cmdDomainDeleteRunE,
}

var cmdDomainUpdate = &cobra.Command{
	Use:   "update",
	Short: "Update a domain. Any field specified here will totally overwrite the current value (won't be appended)",
	RunE:  cmdDomainUpdateRunE,
}

var cmdDomainAssign = &cobra.Command{
	Use:   "assign",
	Short: "Assing a device to a domain",
	RunE:  cmdDomainAssignRunE,
}

var cmdDomainCreate = &cobra.Command{
	Use:   "create",
	Short: "Create a domain. If you dont specify a domain administrator, your address will be set as administrator of the new domain",
	RunE:  cmdDomainCreateRunE,
}

func cmdDomainCreateRunE(cmd *cobra.Command, args []string) error {
	amap, err := parseDomainFlags(cmd)
	if err != nil {
		return err
	}

	return commandCall("idefix", "domain.create", amap, getTimeout(cmd))
}

func cmdDomainAssignRunE(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	address, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
	}

	amap := make(map[string]interface{})
	amap["name"] = name
	amap["address"] = address

	return commandCall("idefix", "domain.assign", amap, getTimeout(cmd))
}

func cmdDomainUpdateRunE(cmd *cobra.Command, args []string) error {
	amap, err := parseDomainFlags(cmd)
	if err != nil {
		return err
	}

	return commandCall("idefix", "domain.update", amap, getTimeout(cmd))
}

func parseDomainFlags(cmd *cobra.Command) (map[string]interface{}, error) {
	amap := make(map[string]interface{})

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return nil, err
	}
	amap["name"] = name

	sallow, err := cmd.Flags().GetString("allow")
	if err != nil {
		return nil, err
	}
	if cmd.Flags().Changed("allow") {
		dummy := make(map[string]interface{})
		err = json.Unmarshal([]byte(sallow), &dummy)
		if err != nil {
			return nil, fmt.Errorf("Cannot parse allow rule: %w", err)
		}
		amap["allow"] = sallow
	}

	sdeny, err := cmd.Flags().GetString("deny")
	if err != nil {
		return nil, err
	}
	if cmd.Flags().Changed("deny") {
		dummy := make(map[string]interface{})
		err = json.Unmarshal([]byte(sdeny), &dummy)
		if err != nil {
			return nil, fmt.Errorf("Cannot parse deny rule: %w", err)
		}
		amap["deny"] = sdeny
	}

	env := make(map[string]interface{})
	senv, err := cmd.Flags().GetString("env")
	if err != nil {
		return nil, err
	}
	if cmd.Flags().Changed("env") {
		err = json.Unmarshal([]byte(senv), &env)
		if err != nil {
			return nil, fmt.Errorf("Cannot parse environment: %w", err)
		}
		amap["env"] = env
	}

	admins, err := cmd.Flags().GetStringArray("admin")
	if err != nil {
		return nil, err
	}
	if cmd.Flags().Changed("admin") {
		amap["admins"] = admins
	}

	return amap, nil
}

func cmdDomainGetRunE(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	amap := make(map[string]interface{})
	amap["name"] = name

	return commandCall("idefix", "domain.get", amap, getTimeout(cmd))
}

func cmdDomainDeleteRunE(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	amap := make(map[string]interface{})
	amap["name"] = name

	fmt.Println("You are about to delete the domain:", name)
	if result, _ := pterm.DefaultInteractiveConfirm.Show(); !result {
		return nil
	}
	if err := commandCall("idefix", "domain.get", amap, getTimeout(cmd)); err != nil {
		return err
	}

	return commandCall("idefix", "domain.delete", amap, getTimeout(cmd))
}
