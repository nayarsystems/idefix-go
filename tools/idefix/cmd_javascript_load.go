package main

import (
	"fmt"
	"os"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var cmdJsLoad = &cobra.Command{
	Use:   "load",
	Short: "Load a new javascript code to the remote module",
	RunE:  cmdJsLoadRunE,
}

func cmdJsLoadRunE(cmd *cobra.Command, args []string) error {
	p, err := getJsParams(cmd)
	if err != nil {
		return err
	}

	codeFile, err := cmd.Flags().GetString("vm-code")
	if err != nil {
		return err
	}

	code, err := os.ReadFile(codeFile)
	if err != nil {
		return err
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start("connecting...")
	defer spinner.Stop()
	defer fmt.Println()

	ic, err := getConnectedClient()
	if err != nil {
		spinner.Fail(err.Error())
		return nil
	}

	spinner.Info(fmt.Sprintf("loading new code to instance '%s'...", p.instance))

	// Create the new instance
	_, err = ic.Call(p.address, &m.Message{
		To: fmt.Sprintf("%s.cmd.load", p.instance),
		Data: map[string]any{
			"code": code,
		},
	}, p.timeout)
	if err != nil {
		spinner.Fail(err.Error())
		return nil
	}
	spinner.Success("")
	return nil
}
