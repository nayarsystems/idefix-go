package main

import (
	"encoding/json"
	"fmt"

	"github.com/nayarsystems/idefix-go/messages"
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
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		format = "pretty"
	}
	if p.UID == "" {
		for p.Continue {
			spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start(fmt.Sprintf(
				"Query for raw events from domain %q, limit: %d, cid: %s, since: %v, for: %d", p.Domain, p.Limit, p.Cid, p.Since, p.Timeout))

			//(p.Domain, p.AddressFilter, p.Since, p.Limit, p.Cid, p.Timeout)
			m, err := ic.EventsGet(&messages.EventsGetMsg{
				Domain:         p.Domain,
				Address:        p.AddressFilter,
				Since:          p.Since,
				Limit:          p.Limit,
				ContinuationID: p.Cid,
				Timeout:        p.Timeout,
				Type:           p.Type,
				NoPayload:      p.NoPayload,
			})
			if err != nil {
				spinner.Fail()
				return err
			}
			spinner.Success()

			if len(m.Events) == 0 {
				return nil
			}

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
			p.Cid = m.ContinuationID
		}
	} else {
		spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start(fmt.Sprintf(
			"Query for raw event: uid: %v, for: %v", p.UID, p.Timeout))

		e, err := ic.GetEventByUID(p.UID, p.Timeout)
		if err != nil {
			spinner.Fail()
			return err
		}
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

	return nil
}
