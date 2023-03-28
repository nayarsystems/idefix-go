package main

import (
	"encoding/json"
	"fmt"

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

	benchmark, _ := cmd.Flags().GetBool("benchmark")

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
					headerSeparatorRune := []rune(header)
					for i := 0; i < len(headerSeparatorRune); i++ {
						headerSeparatorRune[i] = '~'
					}
					headerSeparator := string(headerSeparatorRune)
					fmt.Println(headerSeparator)
					fmt.Println(header)
					fmt.Println(headerSeparator)

					for _, event := range statesSource.States {
						je, err := json.Marshal(event.Delta)
						if err != nil {
							fmt.Println(err)
							continue
						}
						fmt.Printf("%v: %s\n", event.Timestamp, string(je))
					}
					if benchmark {
						idf.BenchmarkBstates(statesSource.States)
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

func RunBenchmark() {

}
