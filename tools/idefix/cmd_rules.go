package main

import (
	"encoding/json"
	"fmt"

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
	amap := make(map[string]interface{})

	amap["address"], err = cmd.Flags().GetString("address")
	if err != nil {
		return err
	}

	return commandCall("idefix", "lrules.get", amap, getTimeout(cmd))
}

func cmdRulesUpdateRunE(cmd *cobra.Command, args []string) (err error) {
	amap := make(map[string]interface{})

	amap["address"], err = cmd.Flags().GetString("address")
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
			return fmt.Errorf("Cannot parse allow rule: %w", err)
		}
		amap["allow"] = sallow
	}

	sdeny, err := cmd.Flags().GetString("deny")
	if err != nil {
		return err
	}
	if cmd.Flags().Changed("deny") {
		dummy := make(map[string]interface{})
		err = json.Unmarshal([]byte(sdeny), &dummy)
		if err != nil {
			return fmt.Errorf("Cannot parse deny rule: %w", err)
		}
		amap["deny"] = sdeny
	}

	return commandCall("idefix", "lrules.update", amap, getTimeout(cmd))
}
