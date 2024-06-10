package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cmdHWTest)
}

var cmdHWTest = &cobra.Command{
	Use:     "hwtest",
	Aliases: []string{"test"},
	Short:   "",
	RunE:    cmdHWTestRunE,
}

func cmdHWTestRunE(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("Error %s", "error")
}
