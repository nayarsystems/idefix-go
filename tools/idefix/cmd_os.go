package main

import (
	"time"

	"github.com/spf13/cobra"
)

func init() {
	cmdOs.PersistentFlags().StringP("address", "a", "", "Device address")
	cmdOs.MarkFlagRequired("address")

	rootCmd.AddCommand(cmdOs)
}

var cmdOs = &cobra.Command{
	Use:   "os",
	Short: "os related commands",
}

type osBaseParams struct {
	address string
	timeout time.Duration
}

func getOsBaseParams(cmd *cobra.Command) (params osBaseParams, err error) {
	params.address, err = cmd.Flags().GetString("address")
	if err != nil {
		return
	}

	timeoutMs, err := cmd.Flags().GetUint("timeout")
	if err != nil {
		return
	}
	params.timeout = time.Duration(timeoutMs) * time.Millisecond
	return
}
