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
