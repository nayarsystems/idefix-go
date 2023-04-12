package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/nayarsystems/idefix/libraries/eval"
	"github.com/spf13/cobra"
)

func init() {
	cmdEventCreate.Flags().StringP("uid", "u", "-", "UID")
	cmdEventCreate.Flags().StringP("schema", "s", "raw", "Payload Schema")
	cmdEventCreate.Flags().StringP("meta", "m", "", "Metadata added to the event in JSON dictionary style")
	cmdEvent.AddCommand(cmdEventCreate)

	cmdEventGet.PersistentFlags().StringP("uid", "u", "", "UID")
	cmdEventGet.PersistentFlags().String("since", "1970", "Query events that happened since this timestamp")
	cmdEventGet.PersistentFlags().Uint("limit", 100, "Limit the number of query results")
	cmdEventGet.PersistentFlags().String("cid", "", "Use a continuationID to get the following results of a previous request")
	cmdEventGet.PersistentFlags().Bool("all", false, "Query all the items")
	cmdEventGet.PersistentFlags().String("timeout", "20s", "If there are no events, wait until some arrive")
	cmdEventGet.PersistentFlags().StringP("address", "a", "", "Filter by the indicated address. Default: get evets from all address in the specified domain")
	cmdEventGet.PersistentFlags().String("meta-filter", "{\"$true\": 1}", "Mongo expression to filter events by the meta field")
	cmdEventGet.PersistentFlags().String("csvdir", "", "Directory path used to export all events in csv format (if specified)")
	cmdEventGet.PersistentFlags().Bool("continue", false, "Perform requests until no events pending")
	cmdEvent.AddCommand(cmdEventGet)

	cmdEventGetRaw.Flags().String("format", "json", "Format to show results: [pretty, json]")
	cmdEventGet.AddCommand(cmdEventGetRaw)

	cmdEventGetBevents.Flags().String("ts-field", "", "Force use of a timestamp field (used when multiple timestamp field are present in the same state). If no field is forced the first found will be used")
	cmdEventGetBevents.Flags().Uint("ts-field-offset", 1970, "Use this year offset when the ts-field specified is a raw numeric field")
	cmdEventGetBevents.Flags().Float32("ts-field-factor", 1, "Use this factor to get milliseconds from the ts-field when it's a raw numeric")
	cmdEventGetBevents.Flags().Bool("benchmark", false, "Perform size efficiency benchmark after getting bstates")
	cmdEventGetBevents.Flags().String("field-match", ".*", "A regex to only show changes on matched fields")
	cmdEventGetBevents.Flags().Bool("field-align", false, "Try to align all field's values in the same column")
	cmdEventGetBevents.Flags().Bool("field-align-hs", false, "Add row separator (used when field-align is true)")
	cmdEventGet.AddCommand(cmdEventGetBevents)

	rootCmd.AddCommand(cmdEvent)
}

var cmdEvent = &cobra.Command{
	Use:     "event",
	Aliases: []string{"events"},
	Short:   "Manage idefix events",
}

var cmdEventGet = &cobra.Command{
	Use:   "get <raw|bstates>",
	Short: "Get events",
}

var cmdEventGetRaw = &cobra.Command{
	Use:   "raw { [DOMAIN] | -u [UID] | -a [ADDRESS] }",
	Short: "Get all kind of events. The events will be shown in raw format",
	RunE:  cmdEventGetRawRunE,
	Args:  cobra.MaximumNArgs(1),
}

var cmdEventGetBevents = &cobra.Command{
	Use:   "bstates { [DOMAIN] | -u [UID] | -a [ADDRESS] }",
	Short: "Get bstates based events",
	RunE:  cmdEventGetBstatesRunE,
	Args:  cobra.MaximumNArgs(1),
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
	uid, _ := cmd.Flags().GetString("uid")
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

	err = ic.SendEvent(payload, schema, tmeta, uid, time.Second)
	if err != nil {
		return err
	}

	return nil
}

type GetEventsBaseParams struct {
	UID           string
	Domain        string
	Limit         uint
	Cid           string
	Timeout       time.Duration
	Since         time.Time
	AddressFilter string
	MetaFilter    eval.CompiledExpr
	Continue      bool
	Csvdir        string
}

func parseGetEventsBaseParams(cmd *cobra.Command, args []string) (*GetEventsBaseParams, error) {
	var err error
	params := &GetEventsBaseParams{}
	if len(args) == 1 {
		params.Domain = args[0]
	}
	params.UID, _ = cmd.Flags().GetString("uid")
	params.Limit, _ = cmd.Flags().GetUint("limit")
	params.Cid, _ = cmd.Flags().GetString("cid")
	timeoutraw, _ := cmd.Flags().GetString("timeout")
	params.Timeout, _ = time.ParseDuration(timeoutraw)
	sinceraw, _ := cmd.Flags().GetString("since")
	params.Continue, _ = cmd.Flags().GetBool("continue")
	if params.Continue {
		params.Limit = 100
	}
	params.Csvdir, _ = cmd.Flags().GetString("csvdir")
	params.Since, err = dateparse.ParseStrict(sinceraw)
	if err != nil {
		return nil, fmt.Errorf("cannot parse 'since': %w", err)
	}
	params.AddressFilter, _ = cmd.Flags().GetString("address")
	metaFilterExpr, _ := cmd.Flags().GetString("meta-filter")
	params.MetaFilter, err = eval.CompileExpr(metaFilterExpr)
	if err != nil {
		return nil, fmt.Errorf("cannot parse 'meta-filter': %w", err)
	}
	return params, nil
}
