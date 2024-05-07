package main

import (
	"time"

	"github.com/spf13/cobra"
)

func init() {
	cmdMenhir.PersistentFlags().StringP("address", "a", "", "Device address")
	cmdMenhir.MarkFlagRequired("address")

	cmdMenhir.PersistentFlags().StringP("instance", "i", "", "menhir instance")
	cmdMenhir.MarkFlagRequired("instance")

	rootCmd.AddCommand(cmdMenhir)
}

var cmdMenhir = &cobra.Command{
	Use:   "menhir",
	Short: "menhir related commands",
}

type menhirParams struct {
	address, instance string
	timeout           time.Duration
}

func getMenhirParams(cmd *cobra.Command) (params menhirParams, err error) {
	params.address, err = cmd.Flags().GetString("address")
	if err != nil {
		return
	}
	params.instance, err = cmd.Flags().GetString("instance")
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
