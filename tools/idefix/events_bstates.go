package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"

	"github.com/jedib0t/go-pretty/v6/table"
	idf "github.com/nayarsystems/idefix-go"
	ie "github.com/nayarsystems/idefix-go/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func cmdEventGetBstatesRunE(cmd *cobra.Command, args []string) error {
	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	var bp *GetEventsBaseParams
	if bp, err = parseGetEventsBaseParams(cmd, args); err != nil {
		return err
	}
	p := idf.GetBstatesParams{
		Domain:        bp.Domain,
		Since:         bp.Since,
		Limit:         bp.Limit,
		Cid:           bp.Cid,
		AddressFilter: bp.AddressFilter,
		MetaFilter:    bp.MetaFilter,
		Timeout:       bp.Timeout,
	}
	p.ForceTsField, _ = cmd.Flags().GetString("ts-field")
	p.RawTsFieldYearOffset, _ = cmd.Flags().GetUint("ts-field-offset")
	p.RawTsFieldFactor, _ = cmd.Flags().GetFloat32("ts-field-factor")
	fieldNameRegexStr, _ := cmd.Flags().GetString("field-match")
	fieldNameRegex, err := regexp.Compile(fieldNameRegexStr)
	if err != nil {
		return err
	}
	benchmark, _ := cmd.Flags().GetBool("benchmark")
	fieldAlign, _ := cmd.Flags().GetBool("field-align")

	keepPolling := true
	res := idf.GetBstatesResult{}
	for keepPolling {
		spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start(fmt.Sprintf("Query for bstates events from domain %q, limit: %d, cid: %s, since: %v, for: %d", p.Domain, p.Limit, p.Cid, p.Since, p.Timeout))
		var sread uint
		sread, p.Cid, err = idf.GetBstates(ic, &p, res)
		timeout := false
		if err != nil {
			if !ie.ErrTimeout.Is(err) {
				spinner.Fail()
				return err
			}
		}
		spinner.Success()
		keepPolling = !timeout && sread == p.Limit && bp.Continue && p.Cid != ""
	}

	for domain, domainMap := range res {
		for address, addressMap := range domainMap {
			for schema, schemaMap := range addressMap {
				for _, statesSource := range schemaMap {
					header := fmt.Sprintf("~~~~~~~~~ DOMAIN: %s, ADDRESS: %s, SCHEMA: %s, META: %s~~~~~~~~~", domain, address, schema, statesSource.MetaRaw)
					printHeader(header)
					var states []*idf.Bstate
					var blobStarts []int
					var blobEnds []int
					for _, blob := range statesSource.Blobs {
						blobStarts = append(blobStarts, len(states))
						blobEnds = append(blobEnds, len(states)+len(blob.States))
						states = append(states, blob.States...)
					}
					var deltas []map[string]interface{}
					deltas, err = idf.GetDeltaStates(states)
					if err != nil {
						fmt.Println("fix me: " + err.Error())
						continue
					}

					fieldsDesc := statesSource.Blobs[0].States[0].State.GetSchema().GetFields()
					dfieldsDesc := statesSource.Blobs[0].States[0].State.GetSchema().GetDecodedFields()
					fieldNames := []string{}
					for _, f := range fieldsDesc {
						fieldNames = append(fieldNames, f.Name)
					}
					for _, f := range dfieldsDesc {
						fieldNames = append(fieldNames, f.Name)
					}
					var matchedFields []string
					if fieldNameRegexStr != ".*" {
						for i := len(deltas) - 1; i >= 0; i-- {
							d := deltas[i]
							newD := map[string]interface{}{}
							for f, v := range d {
								if fieldNameRegex.MatchString(f) {
									newD[f] = v
								}
							}
							deltas[i] = newD
						}
						for _, fname := range fieldNames {
							if fieldNameRegex.MatchString(fname) {
								matchedFields = append(matchedFields, fname)
							}
						}
					} else {
						matchedFields = fieldNames
					}
					sort.Strings(matchedFields)

					for blobIdx, blob := range statesSource.Blobs {
						blobHeader := fmt.Sprintf("~~~~~~~~~ NEW BLOB: UID = %s, DATE: %v, EVENTS: %d ~~~~~~~~~", blob.UID, blob.Timestamp, len(blob.States))
						printHeader(blobHeader)

						t := table.NewWriter()
						t.SetOutputMirror(os.Stdout)
						header := table.Row{"TS"}
						for _, fname := range matchedFields {
							header = append(header, fname)
						}
						t.AppendHeader(header)

						blobDeltas := deltas[blobStarts[blobIdx]:blobEnds[blobIdx]]
						blobStates := states[blobStarts[blobIdx]:blobEnds[blobIdx]]
						if fieldAlign {
							for i, d := range blobDeltas {
								r, err := getEventRow(matchedFields, blobStates[i], d)
								if err != nil {
									fmt.Println(err)
									continue
								}
								t.AppendRow(r)
								t.AppendSeparator()
							}
							t.Render()
						} else {
							for i, s := range blobDeltas {
								je, err := json.Marshal(s)
								if err != nil {
									fmt.Println(err)
									continue
								}
								fmt.Printf("%v: %s\n", blobStates[i].Timestamp, string(je))
							}
						}
						if benchmark {
							//fmt.Printf("\n << BLOB STATS >>\n")
							fmt.Printf("\n\n")
							idf.BenchmarkBstates(blob, blobStates)
							//fmt.Printf("<< BLOB STATS END >>\n\n")
							fmt.Printf("\n\n")
						}
					}
				}
			}
		}
	}
	if p.Cid != "" {
		fmt.Println("CID:", p.Cid)
	} else {
		fmt.Println("no events left")
	}

	return nil
}

func getEventRow(fieldsNames []string, state *idf.Bstate, delta map[string]interface{}) (row table.Row, err error) {
	// colvalues := map[string]string{}
	row = append(row, state.Timestamp)
	for _, fname := range fieldsNames {
		dv, ok := delta[fname]
		if ok {
			row = append(row, dv)
		} else {
			row = append(row, "")
		}
	}
	// 	var fv string
	// 	dv, ok := delta[fname]
	// 	if ok {
	// 		fv = fmt.Sprintf("%v", dv)
	// 	} else {
	// 		dv, err = state.State.Get(fname)
	// 		if err != nil {
	// 			return "", err
	// 		}
	// 		fieldRune := []rune(fmt.Sprintf("%v", dv))
	// 		for i := 0; i < len(fieldRune); i++ {
	// 			fieldRune[i] = ' '
	// 		}
	// 		fv = string(fieldRune)
	// 	}
	// 	colvalues[fname] = fv
	// }
	// for _, fname := range fieldsNames {
	// 	msg += " | "
	// 	v := colvalues[fname]
	// 	msg += fmt.Sprintf("%v", v)
	// }
	return
}

func printHeader(title string) {
	headerSeparatorRune := []rune(title)
	for i := 0; i < len(headerSeparatorRune); i++ {
		headerSeparatorRune[i] = '~'
	}
	headerSeparator := string(headerSeparatorRune)
	fmt.Println(headerSeparator)
	fmt.Println(title)
	fmt.Println(headerSeparator)
}
