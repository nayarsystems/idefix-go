package main

import (
	"encoding/json"
	"fmt"

	idf "github.com/nayarsystems/idefix-go"
	"github.com/spf13/cobra"
)

func init() {
	cmdConfig.AddCommand(cmdConfigLoad)

	cmdConfigStore.Flags().StringP("broker", "b", "ssl://mqtt.terathings.com", "Broker Address")
	cmdConfigStore.Flags().StringP("encoding", "e", "mg", "Encoding")
	cmdConfigStore.Flags().StringP("address", "a", "", "Address")
	cmdConfigStore.Flags().StringP("token", "t", "", "Token")
	cmdConfig.AddCommand(cmdConfigStore)

	rootCmd.AddCommand(cmdConfig)
}

var cmdConfig = &cobra.Command{
	Use:   "config",
	Short: "Manage idefix configurations",
}

var cmdConfigLoad = &cobra.Command{
	Use:   "show <name>",
	Short: "Show and print configuration",
	Args:  cobra.MinimumNArgs(1),
	RunE:  cmdConfigShowRunE,
}

func cmdConfigShowRunE(cmd *cobra.Command, args []string) error {

	cfg, err := idf.ReadConfig(args[0])
	if err != nil {
		return err
	}

	j, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Printf("%s\n", j)

	return nil
}

var cmdConfigStore = &cobra.Command{
	Use:   "store <name>",
	Short: "Store or modify a configuration",
	Args:  cobra.MinimumNArgs(1),
	RunE:  cmdConfigStoreRunE,
}

func cmdConfigStoreRunE(cmd *cobra.Command, args []string) error {
	cfg, err := idf.ReadConfig(args[0])
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("broker") {
		cfg.Broker, _ = cmd.Flags().GetString("broker")
	}

	if cmd.Flags().Changed("encoding") {
		cfg.Encoding, _ = cmd.Flags().GetString("encoding")
	}

	if cmd.Flags().Changed("address") {
		cfg.Address, _ = cmd.Flags().GetString("address")
	}

	if cmd.Flags().Changed("token") {
		cfg.Token, _ = cmd.Flags().GetString("token")
	}

	if err := idf.UpdateConfig(cfg); err != nil {
		return err
	}

	return cmdConfigShowRunE(cmd, args)
}
