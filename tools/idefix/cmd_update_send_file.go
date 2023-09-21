package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	idefixgo "github.com/nayarsystems/idefix-go"
	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func cmdUpdateSendFileRunE(cmd *cobra.Command, args []string) error {
	p, err := getUpdateParams(cmd)
	if err != nil {
		return err
	}

	updatefile, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	p.upgradeType = "bin"
	p.rollbackType = "bin"

	updatebytes, err := os.ReadFile(updatefile)
	if err != nil {
		return err
	}

	upgradeEnvPath, rollbackEnvPath, upgradeBinPath, rollbackBinPath, exitType := getRawParams(p)

	dsthash := Sha256B64(updatebytes)
	upgradeBinPathMsg := upgradeBinPath
	if !filepath.IsAbs(upgradeBinPathMsg) {
		upgradeBinPathMsg = fmt.Sprintf("(relative to idefix binary) %s", upgradeBinPathMsg)
	}
	rollbackBinPathMsg := rollbackBinPath
	if !filepath.IsAbs(rollbackBinPathMsg) {
		rollbackBinPathMsg = fmt.Sprintf("(relative to idefix binary) %s", rollbackBinPathMsg)
	}

	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Upgrade params", ""},
		{"Target", p.targetStr},
		{"Reason", p.reason},
		{"Upgrade file size", KB(uint64(len(updatebytes)))},
		{"Updated file hash", dsthash},
		{"Create rollback (full binary)", fmt.Sprintf("%v", p.createRollback)},
		{"Upgrade path", upgradeBinPathMsg},
		{"Rollback path", rollbackBinPathMsg},
		{"Stability time (s)", fmt.Sprintf("%v", p.stabilitySecs)},
		{"Healthy time (s)", fmt.Sprintf("%v", p.healthySecs)},
		{"Check ppp link", fmt.Sprintf("%v", p.checkPPP)},
		{"Check transport link", fmt.Sprintf("%v", p.checkTransport)},
		{"Stop countdown (s)", fmt.Sprintf("%v", p.stopToutSecs)},
		{"Halt timeout (s)", fmt.Sprintf("%v", p.haltToutSecs)},
	}).Render()

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}

	msg := map[string]interface{}{
		"report": false,
	}
	ret, err := ic.Call(p.address, &m.Message{To: "sys.cmd.info", Data: msg}, time.Second)
	if err != nil {
		return err
	}
	address, version, bootcnt, err := getDevInfoFromMsg(ret.Data)
	if err != nil {
		return err
	}

	// Check space available
	freeSpace, err := idefixgo.GetFree(ic, p.address, "", p.tout)
	if err != nil {
		return err
	}

	fmt.Println()
	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Device", ""},
		{"Address", address},
		{"Boot Counter", fmt.Sprintf("%d", bootcnt)},
		{"Version", version},
		{"Free space", KB(freeSpace)},
	}).Render()

	fmt.Println()

	// if freeSpace < uint64(len(updatebytes)) {
	// 	return fmt.Errorf("not enough space available: update (%d) > free (%d)", len(updatebytes), freeSpace)
	// }

	fmt.Println()

	if result, _ := pterm.DefaultInteractiveConfirm.Show(); !result {
		return nil
	}
	spinner, _ := pterm.DefaultSpinner.WithShowTimer(false).Start("sending upgrade env file...")
	defer spinner.Stop()

	err = sendEnvFile(ic, p.address, p, upgradeEnvPath, false, "", "", p.tout)
	if err != nil {
		return err
	}

	spinner.UpdateText("sending upgrade bin file...")
	receivedHash, err := idefixgo.FileWrite(ic, p.address, upgradeBinPath, updatebytes, 0744, p.tout)
	if err != nil {
		return err
	}
	if receivedHash != dsthash {
		return fmt.Errorf("received upgrade hash (%s) != expected upgrade hash (%s)", receivedHash, dsthash)
	}

	if p.createRollback && p.target == m.IdefixTargetExec {
		spinner.UpdateText("sending rollback env file...")
		err = sendEnvFile(ic, p.address, p, rollbackEnvPath, true, "", "", p.tout)
		if err != nil {
			return err
		}
		spinner.UpdateText("backup idefix file...")
		err = idefixgo.FileCopy(ic, p.address, IdefixExecIdefixRelativePath, rollbackBinPath, p.tout)
		if err != nil {
			return err
		}
	}

	spinner.UpdateText("sending update request...")
	res, err := idefixgo.ExitToUpdate(ic, p.address,
		exitType,
		p.reason,
		time.Duration(p.stopToutSecs)*time.Second,
		time.Duration(p.haltToutSecs)*time.Second,
		p.tout)
	if err != nil {
		return err
	}
	pterm.Success.Printf("Response. %#v\nUpdate completed!\nDevice should reboot now...\n", res)
	return nil
}
