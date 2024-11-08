package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
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

	msg := &m.SysInfoReqMsg{
		Report:          report,
		ReportInstances: reportFilter,
	}
	var info m.SysInfoResMsg
	err = ic.Call2(addr, &m.Message{To: "sys.cmd.info", Data: msg}, &info, getTimeout(cmd))
	if err != nil {
		return fmt.Errorf("cannot get the device info: %w", err)
	}
	var configSyncInfoMsg string
	if info.ConfigInfo.SyncInfo.Msg != "" {
		configSyncInfoMsg = info.ConfigInfo.SyncInfo.Msg
	} else {
		configSyncInfoMsg = info.ConfigInfo.SyncInfo.Error
	}
	cloudFileSha256, err := base64.StdEncoding.DecodeString(info.ConfigInfo.CloudFileSha256)
	if err != nil {
		return fmt.Errorf("cannot decode cloud file sha256: %w", err)
	}

	cloudFileSha256Hex := fmt.Sprintf("%x", cloudFileSha256)

	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Device info", ""},
		{"Address", info.Address},
		{"Product", info.Product},
		{"Board", info.Board},
		{"Boot counter", fmt.Sprintf("%d", info.BootCnt)},
		{"Version", info.Version},
		{"Launcher version", info.LauncherVersion},
		{"Config file path", info.ConfigInfo.CloudFile},
		{"Config file sha256", cloudFileSha256Hex},
		{"Config file sync status", configSyncInfoMsg},
	}).Render()

	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Device status", ""},
		{"Uptime", fmt.Sprintf("%v", info.Uptime)},
		{"Last run uptime", fmt.Sprintf("%v", info.LastRunUptime)},
		{"Last run exit cause", fmt.Sprintf("%v (exit code: %d)", info.LastRunExitCause, info.LastRunExitCode)},
		{"Last run exit issued by", fmt.Sprintf("%v", info.LastRunExitIssuedBy)},
		{"Last run exit issued at", fmt.Sprintf("%v", info.LastRunExitIssuedAt)},
		{"Execs since launcher started", fmt.Sprintf("%v", info.NumExecs)},
	}).Render()

	if info.ConfigInfo.SyncInfo.Error != "" {
		pterm.Warning.Println("There was an error during configuration sync:", info.ConfigInfo.SyncInfo.Error)
	}
	if info.ConfigInfo.CloudFile == "" {
		pterm.Warning.Println("There is no cloud config file configured")
	} else {
		if cloudFileSha256Hex == Sha256Hex([]byte{}) {
			pterm.Warning.Println("The device has virgin state")
		}
	}
	if info.SafeRunExec {
		pterm.Warning.Println("This is a safe run execution")
	}
	if info.ConfigInfo.Dirty {
		pterm.Warning.Println("Configuration is dirty (some extra configuration fragments were loaded)")
	}
	if info.RollbackExec {
		pterm.Warning.Println("This is a rollback execution")
	}
	if info.LauncherErrorMsg != "" {
		pterm.Warning.Println("Last launcher error:", info.LauncherErrorMsg)
	}
	if len(info.Report) > 0 {
		if b, err := json.MarshalIndent(info.Report, "", "  "); err == nil {
			pterm.Info.Printf("Report:\n%s\n", b)
		}
	}

	fmt.Println()
	return nil
}
