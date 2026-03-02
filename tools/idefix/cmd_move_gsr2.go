package main

import (
	_ "embed"
	"fmt"
	"os"
	"time"

	idefixgo "github.com/nayarsystems/idefix-go"
	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

//go:embed fragments/int.json
var intJSON []byte

func init() {
	cmdMove.PersistentFlags().StringP("address", "a", "", "Device address")
	cmdMove.MarkPersistentFlagRequired("address")
	cmdMove.PersistentFlags().Int("exit-code", 204, "exit code for restart")
	cmdMove.PersistentFlags().Int("stop-delay", 10, "stop delay in seconds")
	cmdMove.PersistentFlags().Int("wait-halt-delay", 10, "wait halt delay in seconds")

	cmdMoveInt.Flags().StringP("src", "s", "", "override embedded int.json with a local file")
	cmdMove.AddCommand(cmdMoveInt)
	cmdMove.AddCommand(cmdMoveProd)

	rootCmd.AddCommand(cmdMove)
}

var cmdMove = &cobra.Command{
	Use:   "move-gsr2",
	Short: "Move GSR2 device between integration and production environments",
}

var cmdMoveInt = &cobra.Command{
	Use:   "int",
	Short: "Move device to integration environment",
	RunE:  cmdMoveIntRunE,
}

var cmdMoveProd = &cobra.Command{
	Use:   "prod",
	Short: "Move device to production environment",
	RunE:  cmdMoveProdRunE,
}

func sendExitCmd(ic *idefixgo.Client, address string, cmd *cobra.Command, exitCause string) error {
	exitCode, _ := cmd.Flags().GetInt("exit-code")
	stopDelay, _ := cmd.Flags().GetInt("stop-delay")
	waitHaltDelay, _ := cmd.Flags().GetInt("wait-halt-delay")

	exitMsg := &m.ExitReqMsg{
		ExitCode:      exitCode,
		ExitCause:     exitCause,
		StopDelay:     time.Duration(stopDelay) * time.Second,
		WaitHaltDelay: time.Duration(waitHaltDelay) * time.Second,
	}

	data, err := exitMsg.ToMsi()
	if err != nil {
		return fmt.Errorf("cannot serialize exit message: %w", err)
	}

	timeout := getTimeout(cmd)
	ret, err := ic.Call(address, &m.Message{To: "sys.cmd.exit", Data: data}, timeout)
	if err != nil {
		return fmt.Errorf("cannot send exit command: %w", err)
	}
	if ret.Err != "" {
		return fmt.Errorf("exit command error: %s", ret.Err)
	}

	return nil
}

func cmdMoveIntRunE(cmd *cobra.Command, args []string) error {
	address, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
	}

	timeout := getTimeout(cmd)

	data := intJSON
	srcPath, _ := cmd.Flags().GetString("src")
	if srcPath != "" {
		data, err = os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("cannot read source file: %w", err)
		}
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start(fmt.Sprintf("Moving %s to integration", address))

	ic, err := getConnectedClient()
	if err != nil {
		spinner.Fail()
		return err
	}
	defer ic.Disconnect()

	_, err = idefixgo.FileWrite(ic, address, "config.d/int.json", data, 0644, timeout)
	if err != nil {
		spinner.Fail()
		return fmt.Errorf("cannot write config.d/int.json: %w", err)
	}

	err = sendExitCmd(ic, address, cmd, "move to int")
	if err != nil {
		spinner.Fail()
		return err
	}

	spinner.Success()
	return nil
}

func cmdMoveProdRunE(cmd *cobra.Command, args []string) error {
	address, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
	}

	timeout := getTimeout(cmd)

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start(fmt.Sprintf("Moving %s to production", address))

	ic, err := getConnectedClient()
	if err != nil {
		spinner.Fail()
		return err
	}
	defer ic.Disconnect()

	err = idefixgo.Remove(ic, address, "config.d/int.json", timeout)
	if err != nil {
		spinner.Fail()
		return fmt.Errorf("cannot remove config.d/int.json: %w", err)
	}

	err = sendExitCmd(ic, address, cmd, "move to prod")
	if err != nil {
		spinner.Fail()
		return err
	}

	spinner.Success()
	return nil
}
