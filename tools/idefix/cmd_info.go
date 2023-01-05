package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	idf "gitlab.com/garagemakers/idefix-go"
)

func init() {
	cmdInfo.Flags().StringP("address", "a", "", "Device address")
	cmdInfo.Flags().BoolP("report", "r", false, "Also request a module report")
	cmdInfo.Flags().StringSliceP("report-filter", "f", []string{}, "List of module instances requested to be reported. Empty to request all instances")
	cmdInfo.MarkFlagRequired("address")

	rootCmd.AddCommand(cmdInfo)
}

var cmdInfo = &cobra.Command{
	Use:   "info",
	Short: "Request device info",
	RunE:  cmdInfoRunE,
}

func cmdInfoRunE(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
	}

	report, err := cmd.Flags().GetBool("report")
	if err != nil {
		return err
	}

	reportFilter, err := cmd.Flags().GetStringSlice("report-filter")
	if err != nil {
		return err
	}

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	msg := map[string]interface{}{
		"report":    report,
		"instances": reportFilter,
	}
	info, err := ic.Call(addr, &idf.Message{To: "sys.cmd.info", Data: msg}, getTimeout(cmd))
	if err != nil {
		return fmt.Errorf("cannot get the device info: %w", err)
	}

	if info.Err != nil {
		return fmt.Errorf("unexpected response: %v", info.Err)
	}

	data, ok := info.Data.(map[string]any)
	if !ok {
		return fmt.Errorf("unexpected response: %v", info.Data)
	}

	printMsi(data)

	return nil
}

func printMsi(data map[string]interface{}) {
	level := 1
	for k, v := range data {
		msi, isMsi := v.(map[string]interface{})
		if isMsi {
			fmt.Printf("%s:\n", k)
			_printMsi(level, k, msi)
		} else {
			fmt.Printf("%s: %v\n", k, v)
		}
	}
}

func _printMsi(level int, name string, data map[string]interface{}) {
	prefix := strings.Repeat("  ", level)
	for k, v := range data {
		msi, isMsi := v.(map[string]interface{})
		if isMsi {
			fmt.Printf("%s%s:\n", prefix, k)
			_printMsi(level+1, k, msi)
		} else {
			fmt.Printf("%s%s: %v\n", prefix, k, v)
		}
	}
}
