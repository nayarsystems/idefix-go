package main

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/jaracil/ei"
	"github.com/jedib0t/go-pretty/v6/table"
	idf "github.com/nayarsystems/idefix-go"
	ie "github.com/nayarsystems/idefix-go/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type csvFileCtx struct {
	file   *os.File
	writer *csv.Writer
}

var csvMap map[string]*csvFileCtx = make(map[string]*csvFileCtx)

type getBstatesCmdParams struct {
	apiParams         *idf.GetBstatesParams
	infinitePolling   bool
	fieldNameRegex    *regexp.Regexp
	fieldNameRegexStr string
	benchmark         bool
	fieldAlign        bool
	fieldAlignHs      bool
	hideBlobs         bool
	csvDir            string
	fieldAliasMap     map[string]string
}

func cmdEventGetBstatesRunE(cmd *cobra.Command, args []string) error {
	csvMap = make(map[string]*csvFileCtx)
	defer func() {
		for _, f := range csvMap {
			f.writer.Flush()
			f.file.Close()
		}
	}()
	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	var bp *GetEventsBaseParams
	if bp, err = parseGetEventsBaseParams(cmd, args); err != nil {
		return err
	}
	cmdParams := getBstatesCmdParams{
		apiParams: &idf.GetBstatesParams{
			UID:           bp.UID,
			Domain:        bp.Domain,
			Since:         bp.Since,
			Limit:         bp.Limit,
			Cid:           bp.Cid,
			AddressFilter: bp.AddressFilter,
			MetaFilter:    bp.MetaFilter,
			Timeout:       bp.Timeout,
		},
		infinitePolling: bp.Continue,
		fieldAliasMap:   map[string]string{},
	}

	cmdParams.apiParams.ForceTsField, _ = cmd.Flags().GetString("ts-field")
	cmdParams.apiParams.RawTsFieldYearOffset, _ = cmd.Flags().GetUint("ts-field-offset")
	cmdParams.apiParams.RawTsFieldFactor, _ = cmd.Flags().GetFloat32("ts-field-factor")
	cmdParams.fieldNameRegexStr, _ = cmd.Flags().GetString("field-match")
	cmdParams.fieldNameRegex, err = regexp.Compile(cmdParams.fieldNameRegexStr)
	if err != nil {
		return err
	}
	cmdParams.benchmark, _ = cmd.Flags().GetBool("benchmark")
	cmdParams.fieldAlign, _ = cmd.Flags().GetBool("field-align")
	cmdParams.fieldAlignHs, _ = cmd.Flags().GetBool("field-align-hs")
	cmdParams.hideBlobs, _ = cmd.Flags().GetBool("hide-blobs")
	cmdParams.csvDir, _ = cmd.Flags().GetString("csvdir")

	// Get field aliases from flags (flag field-alias is a list of key=value pairs)
	fieldAliases, _ := cmd.Flags().GetStringSlice("field-alias")
	for _, pairStr := range fieldAliases {
		pair := strings.Split(pairStr, "=")
		if len(pair) != 2 {
			return fmt.Errorf("field-alias: invalid format")
		}
		cmdParams.fieldAliasMap[pair[0]] = pair[1]
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start()
	if bp.UID == "" {
		err = cmdBstatesGetMultipleBlobs(spinner, ic, &cmdParams)
	} else {
		err = cmdBstatesGetSingleBlob(spinner, ic, &cmdParams)
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
	p *getBstatesCmdParams) (err error) {
	spinner.UpdateText(fmt.Sprintf("Query bstates events of blob %s (timeout: %v)", p.apiParams.UID, p.apiParams.Timeout))
	res := idf.GetBstatesResult{}
	_, _, err = idf.GetBstates(ic, p.apiParams, res)
	if err != nil {
		return err
	}
	showEvents(res, p)
	return
}

func cmdBstatesGetMultipleBlobs(
	spinner *pterm.SpinnerPrinter,
	ic *idf.Client,
	p *getBstatesCmdParams) (err error) {

	res := idf.GetBstatesResult{}
	keepPolling := true
	lastCID := ""
	nReq := 0
	var domainText string
	var addressText string
	if p.apiParams.Domain != "" {
		domainText = p.apiParams.Domain
	} else {
		domainText = "*"
	}
	if p.apiParams.AddressFilter != "" {
		addressText = p.apiParams.AddressFilter
	} else {
		addressText = "*"
	}

	var newBlobs uint
	var totalBlobs uint
	for keepPolling {
		spinner.UpdateText(fmt.Sprintf("Query bstates events (domain: %s, address: %s): req. num: %d (timeout: %v, limit: %d, cid: %s, since: %v), new: %d, total: %d", domainText, addressText, nReq, p.apiParams.Timeout, p.apiParams.Limit, p.apiParams.Cid, p.apiParams.Since, newBlobs, totalBlobs))
		newBlobs, p.apiParams.Cid, err = idf.GetBstates(ic, p.apiParams, res)
		if err != nil && !ie.ErrTimeout.Is(err) {
			return err
		}

		if p.infinitePolling {
			showEvents(res, p)
			// Reinitialize res to avoid appending the same events
			res = idf.GetBstatesResult{}
			if p.apiParams.Cid == "" {
				p.apiParams.Cid = lastCID
			}
			lastCID = p.apiParams.Cid
		}

		totalBlobs += newBlobs
		nReq++
		keepPolling = rootctx.Err() == nil && p.infinitePolling
	}

	if !p.infinitePolling {
		showEvents(res, p)
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

func getActualField(p *getBstatesCmdParams, fieldName string) (new string) {
	if alias, ok := p.fieldAliasMap[fieldName]; ok {
		new = alias
	} else {
		new = fieldName
	}
	return
}

func showEvents(
	res idf.GetBstatesResult,
	p *getBstatesCmdParams) {
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
					if p.fieldNameRegexStr != ".*" {
						for i := len(deltas) - 1; i >= 0; i-- {
							d := deltas[i]
							newD := map[string]interface{}{}
							for f, v := range d {
								if p.fieldNameRegex.MatchString(f) {
									newD[f] = v
								}
							}
							deltas[i] = newD
						}
						for _, fname := range fieldNames {
							if p.fieldNameRegex.MatchString(fname) {
								matchedFields = append(matchedFields, fname)
							}
						}
					} else {
						matchedFields = fieldNames
					}

					actualMatchedFields := []string{}
					for _, f := range matchedFields {
						actualMatchedFields = append(actualMatchedFields, getActualField(p, f))
					}
					sort.Strings(actualMatchedFields)
					sortedMatchedFields := make([]string, len(matchedFields))
					for i, f := range matchedFields {
						//get index of f in actualMatchedFields and put it in the same index in sortedMatchedFields
						idx := -1
						for i, af := range actualMatchedFields {
							if af == getActualField(p, f) {
								idx = i
								break
							}
						}
						if idx == -1 {
							idx = i
						}
						sortedMatchedFields[idx] = f
					}
					matchedFields = sortedMatchedFields

					if p.csvDir != "" {
						// Create CSV directory
						if _, err := os.Stat(p.csvDir); os.IsNotExist(err) {
							os.Mkdir(p.csvDir, os.ModePerm)
						}
						schemaIdBytes, err := base64.StdEncoding.DecodeString(schema)
						if err != nil {
							return
						}
						schemaIdHex := fmt.Sprintf("%x", schemaIdBytes)
						// Write CSV files
						fileName := fmt.Sprintf("%s/%s_%s_%s.csv", p.csvDir, domain, address, schemaIdHex[:8])
						// check if file already exists on map
						var csvCtx *csvFileCtx
						var ok bool
						if csvCtx, ok = csvMap[fileName]; !ok {
							csvCtx = &csvFileCtx{}
							csvCtx.file, err = os.Create(fileName)
							if err != nil {
								return
							}
							csvCtx.writer = csv.NewWriter(csvCtx.file)
							csvMap[fileName] = csvCtx
						}
						csvHeader := []string{"TS"}
						for _, fname := range matchedFields {
							csvHeader = append(csvHeader, getActualField(p, fname))
						}
						csvRecords := [][]string{csvHeader}
						for i, d := range deltas {
							if len(d) > 0 {
								r, err := getEventRow(i, matchedFields, states[i], d)
								if err != nil {
									return
								}
								csvRow := []string{ei.N(r[0]).StringZ()}
								for _, v := range r[1:] {
									csvRow = append(csvRow, ei.N(v).StringZ())
								}
								csvRecords = append(csvRecords, csvRow)
							}
						}
						csvCtx.writer.WriteAll(csvRecords)
					}

					if !p.hideBlobs {
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
								header = append(header, getActualField(p, fname))
							}
							t.AppendHeader(header)

							blobDeltas := deltas[blobStarts[blobIdx]:blobEnds[blobIdx]]
							blobStates := states[blobStarts[blobIdx]:blobEnds[blobIdx]]

							if p.fieldAlign {
								for i, d := range blobDeltas {
									if len(d) > 0 {
										r, err := getEventRow(i, matchedFields, blobStates[i], d)
										if err != nil {
											fmt.Println(err)
											continue
										}
										t.AppendRow(r)
										if p.fieldAlignHs {
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
							if p.benchmark {
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
							header = append(header, getActualField(p, fname))
						}
						t.AppendHeader(header)
						if p.fieldAlign {
							for i, d := range deltas {
								if len(d) > 0 {
									r, err := getEventRow(i, matchedFields, states[i], d)
									if err != nil {
										fmt.Println(err)
										continue
									}
									t.AppendRow(r)
									if p.fieldAlignHs {
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
