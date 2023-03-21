package main

import (
	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cmdEnvironment)
}

var cmdEnvironment = &cobra.Command{
	Use:     "environment",
	Aliases: []string{"env"},
	Short:   "Get your environment info",
	RunE:    cmdEnvironmentRunE,
}

func cmdEnvironmentRunE(cmd *cobra.Command, args []string) error {
	return commandCall(m.IdefixCmdPrefix, m.CmdEnvGet, map[string]interface{}{}, getTimeout(cmd))
}
