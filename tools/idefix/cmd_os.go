package main

import (
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
}

func getOsBaseParams(cmd *cobra.Command) (params osBaseParams, err error) {
	params.address, err = cmd.Flags().GetString("address")
	if err != nil {
		return
	}
	return
}
