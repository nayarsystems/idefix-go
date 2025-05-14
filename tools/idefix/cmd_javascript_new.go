package main

import (
	"encoding/json"
	"fmt"
	"os"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var cmdJsNew = &cobra.Command{
	Use:   "new",
	Short: "Create a javascript instance",
	RunE:  cmdJsNewRunE,
}

func cmdJsNewRunE(cmd *cobra.Command, args []string) error {
	p, err := getJsParams(cmd)
	if err != nil {
		return err
	}

	class, err := cmd.Flags().GetString("vm-class")
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("vm-name")
	if err != nil {
		return err
	}

	trackHeap, err := cmd.Flags().GetBool("vm-track-heap")
	if err != nil {
		return err
	}

	var config map[string]any
	configFile, _ := cmd.Flags().GetString("vm-config")
	if len(configFile) > 0 {
		// Read json file
		raw, err := os.ReadFile(configFile)
		if err != nil {
			return err
		}
		err = json.Unmarshal(raw, &config)
		if err != nil {
			return err
		}
	} else {
		config = map[string]any{}
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start("connecting...")
	defer spinner.Stop()
	defer fmt.Println()

	ic, err := getConnectedClient()
	if err != nil {
		spinner.Fail(err.Error())
		return nil
	}

	spinner.Info(fmt.Sprintf("creating new instance '%s'...", p.instance))

	// Create the new instance
	_, err = ic.Call(p.address, &m.Message{
		To: "sys.module.new",
		Data: map[string]any{
			"class":  "javascript",
			"prefix": p.instance,
			"params": map[string]any{
				"js_code_class":       class,
				"js_code_prefix":      name,
				"js_code_params":      []any{config},
				"js_track_heap_usage": trackHeap,
			},
		},
	}, p.timeout)
	if err != nil {
		spinner.Fail(err.Error())
		return nil
	}
	spinner.Success("")
	return nil
}
