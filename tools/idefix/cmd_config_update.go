package main

import (
	"fmt"
	"os"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	cmdConfigUpdate.Flags().StringP("file", "f", "", "New configuration file")
	cmdConfigUpdate.Flags().Bool("sync-now", false, "force the client to sync the configuration immediately")

	cmdConfig.AddCommand(cmdConfigUpdate)
}

var cmdConfigUpdate = &cobra.Command{
	Use:   "update",
	Short: "update remote client configuration",
	RunE:  cmdConfigUpdateRunE,
}

func cmdConfigUpdateRunE(cmd *cobra.Command, args []string) (err error) {
	conf := cmdConfigGetBaseParams(cmd)

	// Read the new configuration file
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}
	forceSync, err := cmd.Flags().GetBool("sync-now")
	if err != nil {
		return err
	}
	if file == "" && !forceSync {
		return fmt.Errorf("at least either a new configuration file or the --sync-now flag must be provided")
	}

	spinner, err := pterm.DefaultSpinner.Start("connecting...")
	if err != nil {
		return err
	}
	defer spinner.Stop()

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()
	spinner.Success("connected")

	if file != "" {
		newConfig, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("cannot read file %s: %w", file, err)
		}

		spinner.UpdateText("uploading configuration file...")
		msg := &m.AddressConfigUpdateMsg{
			Address: conf.address,
			Config:  newConfig,
		}
		var res m.AddressConfigUpdateResponseMsg
		err = ic.Call2(m.IdefixCmdPrefix, &m.Message{To: m.CmdAddressConfigUpdate, Data: msg}, &res, getTimeout(cmd))
		if err != nil {
			return fmt.Errorf("cannot upload client configuration: %w", err)
		}
		spinner.Success("configuration file uploaded successfully")
	}

	if forceSync {
		spinner.UpdateText("force client to sync its configuration...")
		var syncRes m.SyncConfigResponseMsg
		err = ic.Call2(conf.address, &m.Message{To: m.TopicCmdSyncConfig, Data: m.SyncConfigReqMsg{}}, &syncRes, getTimeout(cmd))
		if err != nil {
			return fmt.Errorf("cannot force client to sync its configuration: %w", err)
		}
		spinner.Success(syncRes.Result)
	}
	fmt.Println()
	return nil
}
