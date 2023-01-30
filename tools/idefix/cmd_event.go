package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	be "github.com/nayarsystems/idefix/libraries/bevents"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	cmdEventCreate.Flags().StringP("schema", "s", "raw", "Payload Schema")
	cmdEventCreate.Flags().StringP("meta", "m", "", "Metadata added to the event in JSON dictionary style")
	cmdEvent.AddCommand(cmdEventCreate)

	cmdEventGet.Flags().String("since", "1970", "Query events that happened since this timestamp")
	cmdEventGet.Flags().Uint("limit", 100, "Limit the number of query results")
	cmdEventGet.Flags().String("cid", "", "Use a continuationID to get the following results of a previous request")
	cmdEventGet.Flags().String("format", "json", "Format to show results: [pretty, json]")
	cmdEventGet.Flags().String("timeout", "20s", "If there are no events, wait until some arrive")
	cmdEventGet.Flags().Bool("all", false, "Query all the items")
	cmdEvent.AddCommand(cmdEventGet)

	rootCmd.AddCommand(cmdEvent)
}

var cmdEvent = &cobra.Command{
	Use:     "event",
	Aliases: []string{"events"},
	Short:   "Manage idefix events",
}

var cmdEventGet = &cobra.Command{
	Use:   "get <domain>",
	Short: "Get events from a domain",
	RunE:  cmdEventGetRunE,
	Args:  cobra.ExactArgs(1),
}

var cmdEventCreate = &cobra.Command{
	Use:   "create <payload>",
	Short: "Create an event",
	RunE:  cmdEventCreateRunE,
	Args:  cobra.ExactArgs(1),
}

func cmdEventCreateRunE(cmd *cobra.Command, args []string) error {
	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	payload := strings.Join(args, " ")
	schema, err := cmd.Flags().GetString("schema")

	// Try to guess a schema
	if err != nil {
		schema = "raw"
		tmp := make(map[string]interface{})
		err = json.Unmarshal([]byte(payload), &tmp)
		if err == nil {
			schema = "json-map"
		} else {
			tmp := []interface{}{}
			err = json.Unmarshal([]byte(payload), &tmp)
			if err == nil {
				schema = "json-array"
			}
		}
	}

	tmeta := make(map[string]interface{})
	if cmd.Flags().Changed("meta") {
		meta, _ := cmd.Flags().GetString("meta")
		err = json.Unmarshal([]byte(meta), &tmeta)
		if err != nil {
			return fmt.Errorf("cannot parse 'meta': %w", err)
		}
	}

	err = ic.SendEvent(payload, schema, tmeta, time.Second)
	if err != nil {
		return err
	}

	return nil
}

func cmdEventGetRunE(cmd *cobra.Command, args []string) error {
	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	domain := args[0]
	limit, _ := cmd.Flags().GetUint("limit")
	cid, _ := cmd.Flags().GetString("cid")
	timeoutraw, _ := cmd.Flags().GetString("timeout")
	timeout, _ := time.ParseDuration(timeoutraw)
	sinceraw, _ := cmd.Flags().GetString("since")
	since, err := dateparse.ParseStrict(sinceraw)
	if err != nil {
		return fmt.Errorf("cannot parse 'since': %w", err)
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start(fmt.Sprintf(
		"Query for events from domain %q, limit: %d, cid: %s, since: %v, for: %d", domain, limit, cid, since, timeout))

	m, err := ic.GetEventsByDomain(domain, since, limit, cid, timeout)
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

func getEventsList(schema *be.EventSchema, raw []byte) ([]*be.Event, error) {
	decoder := be.CreateEventQueue(schema)
	err := decoder.Decode([]byte(raw))
	if err != nil {
		return nil, fmt.Errorf("can't decode event: %v", err)
	}
	events, err := decoder.GetEvents()
	if err != nil {
		return nil, fmt.Errorf("can't decode event: %v", err)
	}
	return events, err
}

func getDeltaEvents(events []*be.Event) ([]map[string]interface{}, error) {
	msiEvents, err := be.GetDeltaMsiEvents(events)
	return msiEvents, err
}
