package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	idf "github.com/nayarsystems/idefix-go"
	"github.com/spf13/cobra"
)

var rootctx context.Context
var cancel context.CancelFunc

var rootCmd = &cobra.Command{
	Use:          "idefix",
	Short:        "idefix multi-tool",
	SilenceUsage: true,
}

func handleInterrupts(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt) // SIGINT

	go func() {
		<-sigChan
		cancel()
	}()
}

func main() {
	rootctx, cancel = context.WithCancel(context.Background())
	handleInterrupts(cancel)

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
