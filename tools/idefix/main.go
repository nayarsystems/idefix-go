package main

import (
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"
	idf "gitlab.com/garagemakers/idefix-go"
)

var rootctx context.Context
var cancel context.CancelFunc

var rootCmd = &cobra.Command{
	Use:          "idefix",
	Short:        "idefix multi-tool",
	SilenceUsage: true,
}

func main() {
	rootctx, cancel = context.WithCancel(context.Background())

	rootCmd.PersistentFlags().StringP("config", "c", "default", "idefix-go config file for connection settings")
	rootCmd.PersistentFlags().UintP("timeout", "", 10000, "global timeout in milliseconds")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func getConnectedClient() (*idf.Client, error) {
	configName, err := rootCmd.Flags().GetString("config")
	if err != nil {
		return nil, err
	}

	client, err := idf.NewClientFromFile(rootctx, configName)
	if err != nil {
		return nil, err
	}

	err = client.Connect()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func getTimeout(cmd *cobra.Command) time.Duration {
	timeout := time.Second * 10
	ptimeout, err := cmd.Flags().GetUint("timeout")
	if err == nil {
		timeout = time.Duration(ptimeout * 1000000)
	}
	return timeout
}
