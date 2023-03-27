package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jaracil/ei"
	be "github.com/nayarsystems/bstates"
	"github.com/nayarsystems/idefix-go/messages"
	"github.com/nayarsystems/idefix/libraries/eval"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func cmdEventGetBstatesRunE(cmd *cobra.Command, args []string) error {
	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	var p *getEventsBaseParams
	if p, err = parseGetEventsBaseParams(cmd, args); err != nil {
		return err
	}
	forceTsField, _ := cmd.Flags().GetString("ts-field")
	schemasMap := map[string]*be.StateSchema{}
	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start(fmt.Sprintf(
		"Query for bstates events from domain %q, limit: %d, cid: %s, since: %v, for: %d", p.domain, p.limit, p.cid, p.since, p.timeout))

	m, err := ic.GetEventsByDomain(p.domain, p.since, p.limit, p.cid, p.timeout)
	if err != nil {
		spinner.Fail()
		return err
	}
	spinner.Success()

	//domain -> address -> schema -> meta-hash -> list of States
	stateMap := map[string]map[string]map[string]map[string]*StatesSource{}

	for _, e := range m.Events {
		if p.addressFilter != "" && e.Address != p.addressFilter {
			continue
		}
		if parseEvent, err := evalMeta(e.Meta, p.metaFilter); !parseEvent {
			continue
		} else {
			if err != nil {
				return err
			}
		}
		schemaId, err := messages.BstatesParseSchemaIdFromType(e.Type)
		if err != nil {
			continue
		}
		var blob []byte
		err = nil
		rawMsi, ok := e.Payload.(map[string]interface{})
		if ok {
			blobI, ok := rawMsi["Data"]
			if !ok {
				fmt.Println("no 'Data' field found")
				continue
			}

			switch v := blobI.(type) {
			case []byte:
				blob = v
			case string:
				var derr error
				blob, derr = base64.StdEncoding.DecodeString(v)
				if derr != nil {
					err = fmt.Errorf("'Data' is a string but is not valid base64: %v", derr)
				}
			default:
				err = fmt.Errorf("can't get a buffer from 'Data' field")
			}
			if err != nil {
				fmt.Printf("%v\n", err)
				continue
			}
		} else {
			b64Str, ok := e.Payload.(string)
			if !ok {
				fmt.Println("wrong payload format")
			}
			var derr error
			blob, derr = base64.StdEncoding.DecodeString(b64Str)
			if derr != nil {
				fmt.Printf("payload is a string but is not a valid base64: %v", derr)
				continue
			}
		}

		var schema *be.StateSchema
		if schema, ok = schemasMap[schemaId]; !ok {
			schemaMsg, err := ic.GetSchema(schemaId, time.Second)
			if err != nil {
				fmt.Printf("schema '%s' was not found: %v\n", schemaId, err)
				continue
			}
			schema = &be.StateSchema{}
			err = schema.UnmarshalJSON([]byte(schemaMsg.Payload))
			if err != nil {
				fmt.Printf("can't parse schema '%s': %v\n", schemaId, err)
				continue
			}
			schemasMap[schemaId] = schema
		}

		var domainMap map[string]map[string]map[string]*StatesSource
		if domainMap, ok = stateMap[e.Domain]; !ok {
			domainMap = map[string]map[string]map[string]*StatesSource{}
			stateMap[e.Domain] = domainMap
		}

		var addressMap map[string]map[string]*StatesSource
		if addressMap, ok = domainMap[e.Address]; !ok {
			addressMap = map[string]map[string]*StatesSource{}
			domainMap[e.Address] = addressMap
		}

		var schemaMap map[string]*StatesSource
		if schemaMap, ok = addressMap[schemaId]; !ok {
			schemaMap = map[string]*StatesSource{}
			addressMap[schemaId] = schemaMap
		}

		metaRaw, err := json.Marshal(e.Meta)
		if err != nil {
			fmt.Printf("can't get raw meta: %v\n", err)
			continue
		}
		metaHashRaw := sha256.Sum256(metaRaw)
		metaHash := base64.StdEncoding.EncodeToString(metaHashRaw[:])

		states, err := getStatesList(schema, blob)
		if err != nil {
			fmt.Printf("can't get states list: %v\n", err)
			continue
		}
		stateSource, ok := schemaMap[metaHash]
		if !ok {
			stateSource = &StatesSource{
				Meta:    e.Meta,
				MetaRaw: string(metaRaw),
				States:  states,
			}
			schemaMap[metaHash] = stateSource
		} else {
			stateSource.States = append(stateSource.States, states...)
		}

		// get timestamp field
		dfields := schema.GetDecodedFields()
		tsFieldName := ""
		for _, f := range dfields {
			if f.Decoder.Name() == be.NumberToUnixTsMsDecoderType {
				if forceTsField == "" {
					tsFieldName = f.Name
					break
				} else {
					if forceTsField == f.Name {
						tsFieldName = f.Name
						break
					}
				}
			}
		}
		stateSource.TimestampField = tsFieldName
	}

	for domain, domainMap := range stateMap {
		for address, addressMap := range domainMap {
			for schema, schemaMap := range addressMap {
				for _, statesSource := range schemaMap {
					header := fmt.Sprintf("~~~~~~~~~ DOMAIN: %s, ADDRESS: %s, SCHEMA: %s, META: %s~~~~~~~~~", domain, address, schema, statesSource.MetaRaw)
					headerSeparatorRune := []rune(header)
					for i := 0; i < len(headerSeparatorRune); i++ {
						headerSeparatorRune[i] = '~'
					}
					headerSeparator := string(headerSeparatorRune)
					fmt.Println(headerSeparator)
					fmt.Println(header)
					fmt.Println(headerSeparator)
					events, err := getDeltaStates(statesSource.States)
					if err != nil {
						fmt.Println(err)
						continue
					}
					for i, event := range events {
						je, err := json.Marshal(event)
						if err != nil {
							fmt.Println(err)
							continue
						}
						if statesSource.TimestampField != "" {
							tsmsI, err := statesSource.States[i].Get(statesSource.TimestampField)
							if err != nil {
								fmt.Printf("fix me: cannot get '%s' value: %v\n", statesSource.TimestampField, err)
								fmt.Println(string(je))
								continue
							}
							tsms, err := ei.N(tsmsI).Int64()
							if err != nil {
								fmt.Printf("fix me: cannot get '%s' value: %v\n", statesSource.TimestampField, err)
								fmt.Println(string(je))
								continue
							}
							ts := time.UnixMilli(tsms)
							fmt.Printf("%v: %s\n", ts, string(je))
						} else {
							fmt.Println(string(je))
						}
					}
				}
			}
		}
	}

	fmt.Println("CID:", m.ContinuationID)
	return nil
}

func evalMeta(meta map[string]interface{}, expr eval.CompiledExpr) (bool, error) {
	metaEnv := map[string]interface{}{}
	for k, v := range meta {
		metaEnv[k] = v
	}
	res := eval.EvalCompiled(expr, metaEnv)
	switch res.Res {
	case eval.ResInvalidExpr, eval.ResInvalidOp, eval.ResTypeMismatch:
		return false, fmt.Errorf("filter expression error (%d): %s", res.Res, res.Iden)
	}
	if res.Res == eval.ResOK {
		return true, nil
	}
	return false, nil
}

type StatesSource struct {
	Meta           map[string]interface{}
	MetaRaw        string
	TimestampField string
	States         []*be.State
}
