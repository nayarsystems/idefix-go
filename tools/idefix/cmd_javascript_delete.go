package main

import (
	"fmt"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var cmdJsDelete = &cobra.Command{
	Use:   "delete",
	Short: "Delete a javascript instance",
	RunE:  cmdJsDeleteRunE,
}

func cmdJsDeleteRunE(cmd *cobra.Command, args []string) error {
	p, err := getJsParams(cmd)
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

	spinner.Info(fmt.Sprintf("deleting instance '%s'...", p.instance))

	// Create the new instance
	_, err = ic.Call(p.address, &m.Message{
		To: "sys.module.delete",
		Data: map[string]any{
			"prefix": p.instance,
		},
	}, p.timeout)
	if err != nil {
		spinner.Fail(err.Error())
		return nil
	}
	spinner.Success("")
	return nil
}
