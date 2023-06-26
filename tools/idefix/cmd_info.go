package main

import (
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
	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Device info", ""},
		{"Address", info.Address},
		{"Product", info.Product},
		{"Board", info.Board},
		{"Boot counter", fmt.Sprintf("%d", info.BootCnt)},
		{"Version", info.Version},
		{"Launcher version", info.LauncherVersion},
	}).Render()

	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Device status", ""},
		{"Uptime", fmt.Sprintf("%v", info.Uptime)},
		{"Last run uptime", fmt.Sprintf("%v", info.LastRunUptime)},
		{"Last run exit cause", fmt.Sprintf("%v (exit code: %d)", info.LastRunExitCause, info.LastRunExitCode)},
		{"Last run exit issued by", fmt.Sprintf("%v", info.LastRunExitIssuedBy)},
		{"Last run exit issued at", fmt.Sprintf("%v", info.LastRunExitIssuedAt)},
		{"Execs since system boot", fmt.Sprintf("%v", info.NumExecs)},
	}).Render()

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
