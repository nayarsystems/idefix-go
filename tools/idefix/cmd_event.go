package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	cmdEventCreate.Flags().StringP("schema", "s", "raw", "Payload Schema")
	cmdEventCreate.Flags().StringP("meta", "m", "", "Metadata added to the event in JSON dictionary style")
	cmdEvent.AddCommand(cmdEventCreate)

	cmdEventGet.Flags().StringP("since", "", "1970", "Query events that happened since this timestamp")
	cmdEventGet.Flags().UintP("limit", "", 100, "Limit the number of query results")
	cmdEventGet.Flags().UintP("skip", "", 0, "Skip elements from the query results")
	cmdEventGet.Flags().StringP("format", "", "json", "Format to show results: [pretty, json]")
	cmdEventGet.Flags().BoolP("all", "", false, "Query all the items")
	cmdEventGet.Flags().BoolP("reverse", "", false, "Show newer results first")
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
			return fmt.Errorf("Cannot parse 'meta': %w", err)
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
	skip, _ := cmd.Flags().GetUint("skip")
	limit, _ := cmd.Flags().GetUint("limit")
	reverse, _ := cmd.Flags().GetBool("reverse")
	sinceraw, _ := cmd.Flags().GetString("since")
	since, err := dateparse.ParseStrict(sinceraw)
	if err != nil {
		return fmt.Errorf("Cannot parse 'since': %w", err)
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start(fmt.Sprintf(
		"Query for events from domain %q, limit: %d, skip: %d, since: %v", domain, limit, skip, since))

	m, err := ic.GetEventsByDomain(domain, since, limit, skip, reverse, time.Second)
	if err != nil {
		spinner.Fail()
		return err
	}
	spinner.Success()

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		format = "pretty"
	}

	for _, e := range m {
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