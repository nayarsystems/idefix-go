package main

import (
	"context"
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
		UID:           bp.UID,
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
	fieldAlignHs, _ := cmd.Flags().GetBool("field-align-hs")
	hideBlobs, _ := cmd.Flags().GetBool("hide-blobs")
	csvDir, _ := cmd.Flags().GetString("csvdir")
	if csvDir != "" {
		return fmt.Errorf("csvdir: not implemented")
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start()
	if bp.UID == "" {
		err = cmdBstatesGetMultipleBlobs(spinner, ic, bp, p, fieldNameRegex, fieldNameRegexStr, benchmark, fieldAlign, fieldAlignHs, hideBlobs)
	} else {
		err = cmdBstatesGetSingleBlob(spinner, ic, bp, p, fieldNameRegex, fieldNameRegexStr, benchmark, fieldAlign, fieldAlignHs, hideBlobs)
	}
	if err != nil {
		spinner.Fail()
	} else {
		spinner.Success()
	}
	return err
}

func cmdBstatesGetSingleBlob(
	spinner *pterm.SpinnerPrinter,
	ic *idf.Client,
	bp *GetEventsBaseParams,
	p idf.GetBstatesParams,
	fieldNameRegex *regexp.Regexp,
	fieldNameRegexStr string,
	benchmark bool,
	fieldAlign bool,
	fieldAlignHs bool,
	hideBlobs bool) (err error) {
	spinner.UpdateText(fmt.Sprintf("Query bstates events of blob %s (timeout: %v)", p.UID, p.Timeout))
	res := idf.GetBstatesResult{}
	_, _, err = idf.GetBstates(ic, &p, res)
	if err != nil {
		return err
	}
	showEvents(res, bp, fieldNameRegex, fieldNameRegexStr, benchmark, fieldAlign, fieldAlignHs, hideBlobs)
	return
}

func cmdBstatesGetMultipleBlobs(
	spinner *pterm.SpinnerPrinter,
	ic *idf.Client,
	bp *GetEventsBaseParams,
	p idf.GetBstatesParams,
	fieldNameRegex *regexp.Regexp,
	fieldNameRegexStr string,
	benchmark bool,
	fieldAlign bool,
	fieldAlignHs bool,
	hideBlobs bool) (err error) {

	res := idf.GetBstatesResult{}
	keepPolling := true
	lastCID := ""
	nReq := 0
	var domainText string
	var addressText string
	if bp.Domain != "" {
		domainText = bp.Domain
	} else {
		domainText = "*"
	}
	if bp.AddressFilter != "" {
		addressText = bp.AddressFilter
	} else {
		addressText = "*"
	}

	var newBlobs uint
	var totalBlobs uint
	for keepPolling {
		spinner.UpdateText(fmt.Sprintf("Query bstates events (domain: %s, address: %s): req. num: %d (timeout: %v, limit: %d, cid: %s, since: %v), new: %d, total: %d", domainText, addressText, nReq, p.Timeout, p.Limit, p.Cid, p.Since, newBlobs, totalBlobs))
		newBlobs, p.Cid, err = idf.GetBstates(ic, &p, res)
		if err != nil && !ie.ErrTimeout.Is(err) {
			return err
		}

		if bp.Continue {
			showEvents(res, bp, fieldNameRegex, fieldNameRegexStr, benchmark, fieldAlign, fieldAlignHs, hideBlobs)
			// Reinitialize res to avoid appending the same events
			res = idf.GetBstatesResult{}
			if p.Cid == "" {
				p.Cid = lastCID
			}
			lastCID = p.Cid
		}

		totalBlobs += newBlobs
		nReq++
		keepPolling = rootctx.Err() == nil && bp.Continue
	}

	if !bp.Continue {
		showEvents(res, bp, fieldNameRegex, fieldNameRegexStr, benchmark, fieldAlign, fieldAlignHs, hideBlobs)
	}
	if rootctx.Err() == context.Canceled {
		return nil
	}
	return rootctx.Err()
}

func getEventRow(stateIdx int, fieldsNames []string, state *idf.Bstate, delta map[string]interface{}) (row table.Row, err error) {
	row = append(row, state.Timestamp)
	if stateIdx > 0 {
		for _, fname := range fieldsNames {
			dv, ok := delta[fname]
			if ok {
				row = append(row, dv)
			} else {
				row = append(row, "")
			}
		}
	} else {
		for _, fname := range fieldsNames {
			sv, _ := state.State.Get(fname)
			row = append(row, sv)
		}
	}
	return
}

func showEvents(
	res idf.GetBstatesResult,
	p *GetEventsBaseParams,
	fieldNameRegex *regexp.Regexp,
	fieldNameRegexStr string,
	benchmark bool,
	fieldAlign bool,
	fieldAlignHs bool,
	hideBlobs bool) {
	var err error
	for domain, domainMap := range res {
		for address, addressMap := range domainMap {
			for schema, schemaMap := range addressMap {
				for _, statesSource := range schemaMap {
					sourceHeader := table.NewWriter()
					sourceHeader.SetOutputMirror(os.Stdout)
					sourceHeader.AppendHeader(table.Row{"DOMAIN", "ADDRESS", "SCHEMA", "META", "BLOBS", "EVENTs"})

					var states []*idf.Bstate
					var blobStarts []int
					var blobEnds []int
					for _, blob := range statesSource.Blobs {
						blobStarts = append(blobStarts, len(states))
						blobEnds = append(blobEnds, len(states)+len(blob.States))
						states = append(states, blob.States...)
					}

					sourceHeader.AppendRow(table.Row{domain, address, schema, statesSource.MetaRaw, len(statesSource.Blobs), len(states)})
					sourceHeader.Render()

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

					if !hideBlobs {
						for blobIdx, blob := range statesSource.Blobs {
							bh := table.NewWriter()
							bh.SetOutputMirror(os.Stdout)
							bh.AppendHeader(table.Row{"BLOB UID", "DATE", "EVENT COUNT"})
							bh.AppendRow(table.Row{blob.UID, blob.Timestamp, len(blob.States)})
							bh.Render()

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
									if len(d) > 0 {
										r, err := getEventRow(i, matchedFields, blobStates[i], d)
										if err != nil {
											fmt.Println(err)
											continue
										}
										t.AppendRow(r)
										if fieldAlignHs {
											t.AppendSeparator()
										}
									}
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
					} else {
						t := table.NewWriter()
						t.SetOutputMirror(os.Stdout)
						header := table.Row{"TS"}
						for _, fname := range matchedFields {
							header = append(header, fname)
						}
						t.AppendHeader(header)
						if fieldAlign {
							for i, d := range deltas {
								if len(d) > 0 {
									r, err := getEventRow(i, matchedFields, states[i], d)
									if err != nil {
										fmt.Println(err)
										continue
									}
									t.AppendRow(r)
									if fieldAlignHs {
										t.AppendSeparator()
									}
								}
							}
							t.Render()
						} else {
							for i, s := range deltas {
								je, err := json.Marshal(s)
								if err != nil {
									fmt.Println(err)
									continue
								}
								fmt.Printf("%v: %s\n", states[i].Timestamp, string(je))
							}
						}
					}

				}
			}
		}
	}
}
