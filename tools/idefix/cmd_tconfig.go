package main

import (
	"encoding/json"
	"fmt"

	idf "github.com/nayarsystems/idefix-go"
	"github.com/spf13/cobra"
)

func init() {
	cmdToolConfig.AddCommand(cmdToolConfigLoad)

	cmdToolConfigStore.Flags().StringP("broker", "b", "ssl://mqtt.terathings.com", "Broker Address")
	cmdToolConfigStore.Flags().StringP("encoding", "e", "mg", "Encoding")
	cmdToolConfigStore.Flags().StringP("address", "a", "", "Address")
	cmdToolConfigStore.Flags().StringP("token", "t", "", "Token")
	cmdToolConfig.AddCommand(cmdToolConfigStore)

	rootCmd.AddCommand(cmdToolConfig)
}

var cmdToolConfig = &cobra.Command{
	Use:   "tconfig",
	Short: "Manage idefix configurations",
}

var cmdToolConfigLoad = &cobra.Command{
	Use:   "show <name>",
	Short: "Show and print configuration",
	Args:  cobra.MinimumNArgs(1),
	RunE:  cmdToolConfigShowRunE,
}

func cmdToolConfigShowRunE(cmd *cobra.Command, args []string) error {

	cfg, err := idf.ReadConfig(args[0])
	if err != nil {
		return err
	}

	j, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Printf("%s\n", j)

	return nil
}

var cmdToolConfigStore = &cobra.Command{
	Use:   "store <name>",
	Short: "Store or modify a configuration",
	Args:  cobra.MinimumNArgs(1),
	RunE:  cmdToolConfigStoreRunE,
}

func cmdToolConfigStoreRunE(cmd *cobra.Command, args []string) error {
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

	return cmdToolConfigShowRunE(cmd, args)
}
