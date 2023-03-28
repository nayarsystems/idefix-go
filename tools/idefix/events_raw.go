package main

import (
	"encoding/json"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func cmdEventGetRawRunE(cmd *cobra.Command, args []string) error {
	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	var p *GetEventsBaseParams
	if p, err = parseGetEventsBaseParams(cmd, args); err != nil {
		return err
	}
	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start(fmt.Sprintf(
		"Query for raw events from domain %q, limit: %d, cid: %s, since: %v, for: %d", p.Domain, p.Limit, p.Cid, p.Since, p.Timeout))

	m, err := ic.GetEventsByDomain(p.Domain, p.Since, p.Limit, p.Cid, p.Timeout)
	if err != nil {
		spinner.Fail()
		return err
	}
	spinner.Success()

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		format = "pretty"
	}

	// TODO pretty printing

	for _, e := range m.Events {
		switch format {
		case "json":
			je, err := json.Marshal(e)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(string(je))
		default:
			fmt.Printf("%s\n", e.String())
		}
	}

	fmt.Println("CID:", m.ContinuationID)
	return nil
}
