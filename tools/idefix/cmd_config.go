package main

import (
	"context"

	"github.com/spf13/cobra"
)

func init() {
	cmdConfig.PersistentFlags().StringP("address", "a", "", "Device address")
	cmdConfig.MarkPersistentFlagRequired("address")
	rootCmd.AddCommand(cmdConfig)
}

var cmdConfig = &cobra.Command{
	Use:               "config",
	Short:             "Manage a client configuration",
	Args:              cobra.MinimumNArgs(1),
	PersistentPreRunE: configPreRun,
}

type baseConfigParams struct {
	address string
}

type ctxKey string

func cmdConfigGetBaseParams(cmd *cobra.Command) (conf baseConfigParams) {
	if v := cmd.Context().Value(ctxKey("baseConfigParams")); v != nil {
		conf = v.(baseConfigParams)
	}
	return
}

func configPreRun(cmd *cobra.Command, args []string) (err error) {
	var conf baseConfigParams
	conf.address, err = cmd.Flags().GetString("address")
	if err != nil {
		return
	}
	cmd.SetContext(context.WithValue(cmd.Context(), ctxKey("baseConfigParams"), conf))
	return
}
