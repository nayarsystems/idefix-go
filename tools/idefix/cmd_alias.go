package main

import (
	"context"

	"github.com/spf13/cobra"
)

func init() {
	cmdAlias.PersistentFlags().StringP("address", "a", "", "Device address")
	cmdAlias.MarkPersistentFlagRequired("address")
	rootCmd.AddCommand(cmdAlias)
}

var cmdAlias = &cobra.Command{
	Use:               "alias",
	Short:             "Manage client aliases",
	Args:              cobra.MinimumNArgs(1),
	PersistentPreRunE: aliasPreRun,
}

type baseAliasParams struct {
	address string
}

func cmdAliasParams(cmd *cobra.Command) (conf baseAliasParams) {
	if v := cmd.Context().Value(ctxKey("baseConfigParams")); v != nil {
		conf = v.(baseAliasParams)
	}
	return
}

func aliasPreRun(cmd *cobra.Command, args []string) (err error) {
	var conf baseAliasParams
	conf.address, err = cmd.Flags().GetString("address")
	if err != nil {
		return
	}
	cmd.SetContext(context.WithValue(cmd.Context(), ctxKey("baseConfigParams"), conf))
	return
}
