package main

import (
	"time"

	"github.com/spf13/cobra"
)

func init() {
	cmdJs.PersistentFlags().StringP("address", "a", "", "device address")
	cmdJs.PersistentFlags().StringP("instance", "i", "", "javascript instance name")
	cmdJs.MarkPersistentFlagRequired("address")
	cmdJs.MarkPersistentFlagRequired("instance")

	cmdJsNew.Flags().String("vm-class", "", "virtual instance class")
	cmdJsNew.Flags().String("vm-name", "", "virtual instance name")
	cmdJsNew.Flags().String("vm-config", "", "virtual instance configuration file (json object)")
	cmdJsNew.Flags().Bool("vm-track-heap", false, "track heap usage")
	cmdJsNew.MarkFlagRequired("vm-class")
	cmdJsNew.MarkFlagRequired("vm-name")
	cmdJs.AddCommand(cmdJsNew)

	cmdJs.AddCommand(cmdJsDelete)

	cmdJsLoad.Flags().String("vm-code", "", "javascript code file to load")
	cmdJsLoad.MarkFlagRequired("vm-code")
	cmdJs.AddCommand(cmdJsLoad)

	rootCmd.AddCommand(cmdJs)
}

var cmdJs = &cobra.Command{
	Use:   "js",
	Short: "Manage javascript remote module",
}

type jsParams struct {
	timeout  time.Duration
	address  string
	instance string
}

func getJsParams(cmd *cobra.Command) (p *jsParams, err error) {
	p = &jsParams{}

	timeoutMs, err := cmd.Flags().GetUint("timeout")
	if err != nil {
		return
	}
	p.timeout = time.Duration(timeoutMs) * time.Millisecond

	p.address, err = cmd.Flags().GetString("address")
	if err != nil {
		return
	}

	p.instance, err = cmd.Flags().GetString("instance")
	if err != nil {
		return
	}
	return
}
