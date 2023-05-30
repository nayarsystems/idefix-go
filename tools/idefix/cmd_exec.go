package main

import (
	"fmt"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	cmdExec.Flags().StringP("address", "a", "", "Device address")
	cmdExec.Flags().String("cmd", "", "Command to be executed")
	cmdExec.MarkFlagRequired("address")

	rootCmd.AddCommand(cmdExec)
}

var cmdExec = &cobra.Command{
	Use:   "exec",
	Short: "Request device info",
	RunE:  cmdExecRunE,
}

func cmdExecRunE(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
	}

	command, err := cmd.Flags().GetString("cmd")
	if err != nil {
		return err
	}

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	msg := &m.ExecReqMsg{
		Cmd: command,
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start(fmt.Sprintf("requesting command \"%s\"...", command))

	res := &m.ExecResMsg{}

	err = ic.Call2(addr, &m.Message{To: m.TopicCmdExec, Data: msg}, &res, getTimeout(cmd))
	if err != nil {
		spinner.Fail()
		return fmt.Errorf("cannot run exec command: %w", err)
	}
	spinner.Success()

	fmt.Printf("code: %d; success: %v\n", res.Code, res.Success)
	if res.Stdout != "" {
		fmt.Printf("------\nstdout\n------\n%s\n", res.Stdout)
	}

	if res.Stderr != "" {
		fmt.Printf("------\nstderr\n------\n%s\n", res.Stderr)
	}

	return nil
}
