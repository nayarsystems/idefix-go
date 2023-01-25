package main

import (
	"encoding/base64"
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

	cmdEventGet.Flags().StringP("since", "", "1970", "Query events that happened since this timestamp")
	cmdEventGet.Flags().UintP("limit", "", 100, "Limit the number of query results")
	cmdEventGet.Flags().UintP("skip", "", 0, "Skip elements from the query results")
	cmdEventGet.Flags().StringP("format", "", "json", "Format to show results: [pretty, json]")
	cmdEventGet.Flags().BoolP("all", "", false, "Query all the items")
	cmdEventGet.Flags().BoolP("reverse", "", false, "Show newer results first")
	cmdEventGet.Flags().StringP("schemas", "", "", "file containing the schema used to decode payload data. Payload will be shown using its raw format if its schema is not found in this file")
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
	skip, _ := cmd.Flags().GetUint("skip")
	limit, _ := cmd.Flags().GetUint("limit")
	reverse, _ := cmd.Flags().GetBool("reverse")
	sinceraw, _ := cmd.Flags().GetString("since")
	since, err := dateparse.ParseStrict(sinceraw)
	if err != nil {
		return fmt.Errorf("cannot parse 'since': %w", err)
	}

	schemasMap := map[string]*be.EventSchema{}

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

	if false {
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

	//domain -> address -> schema -> list of events
	eventMap := map[string]map[string]map[string][]*be.Event{}

	for _, e := range m {
		rawMsi, ok := e.Payload.(map[string]interface{})
		if !ok {
			return fmt.Errorf("not an msi")
		}
		blobI, ok := rawMsi["Data"]
		if !ok {
			return fmt.Errorf("no Data in msi")
		}
		blobStr, ok := blobI.(string)
		if !ok {
			return fmt.Errorf("data is not a string")
		}
		raw, err := base64.StdEncoding.DecodeString(blobStr)
		if err != nil {
			return fmt.Errorf("blob is not in base64: %v", err)
		}

		var schema *be.EventSchema
		if schema, ok = schemasMap[e.Schema]; !ok {
			schemaMsg, err := ic.GetSchema(e.Schema, time.Second)
			if err != nil {
				return err
			}
			schema = &be.EventSchema{}
			err = schema.UnmarshalJSON([]byte(schemaMsg.Payload))
			if err != nil {
				return err
			}
		}
		blobEvents, err := getEventsList(schema, raw)
		if err != nil {
			return err
		}
		var domainMap map[string]map[string][]*be.Event
		if domainMap, ok = eventMap[e.Domain]; !ok {
			domainMap = map[string]map[string][]*be.Event{}
			eventMap[e.Domain] = domainMap
		}

		var addressMap map[string][]*be.Event
		if addressMap, ok = domainMap[e.Address]; !ok {
			addressMap = map[string][]*be.Event{}
			domainMap[e.Address] = addressMap
		}

		var schemaEvents []*be.Event
		if schemaEvents, ok = addressMap[e.Schema]; !ok {
			schemaEvents = []*be.Event{}
		}
		schemaEvents = append(schemaEvents, blobEvents...)
		addressMap[e.Schema] = schemaEvents
	}

	for domain, domainMap := range eventMap {
		for address, addressMap := range domainMap {
			for schema, schemaEvents := range addressMap {
				header := fmt.Sprintf("~~~~~~~~~ DOMAIN: %s, ADDRESS: %s, SCHEMA: %s ~~~~~~~~~", domain, address, schema)
				headerSeparatorRune := []rune(header)
				for i := 0; i < len(headerSeparatorRune); i++ {
					headerSeparatorRune[i] = '~'
				}
				headerSeparator := string(headerSeparatorRune)
				fmt.Println(headerSeparator)
				fmt.Println(header)
				fmt.Println(headerSeparator)
				deltaEvents, err := getDeltaEvents(schemaEvents)
				if err != nil {
					fmt.Println(err)
					continue
				}
				for _, deltaEvent := range deltaEvents {
					je, err := json.Marshal(deltaEvent)
					if err != nil {
						fmt.Println(err)
						continue
					}
					fmt.Println(string(je))
				}
			}
		}
	}
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
