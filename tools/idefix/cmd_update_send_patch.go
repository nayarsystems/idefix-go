package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jaracil/ei"
	idefixgo "github.com/nayarsystems/idefix-go"
	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/nayarsystems/idefix-go/normalize"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func cmdUpdateSendPatchRunE(cmd *cobra.Command, args []string) error {
	p, err := getUpdateParams(cmd)
	if err != nil {
		return err
	}

	p.upgradeType = "patch"

	p.rollbackType, err = cmd.Flags().GetString("rollback-type")
	if err != nil {
		return err
	}
	if p.rollbackType != "patch" && p.rollbackType != "bin" {
		return fmt.Errorf("invalid rollback type: %s", p.rollbackType)
	}

	patchfile, err := cmd.Flags().GetString("patch")
	if err != nil {
		return err
	}

	patchbytes, err := os.ReadFile(patchfile)
	if err != nil {
		return err
	}

	patchmap := make(map[string]interface{})
	if err := json.Unmarshal(patchbytes, &patchmap); err != nil {
		return err
	}

	normalize.DecodeTypes(patchmap)

	srchashRaw, err := ei.N(patchmap).M("srchash").Bytes()
	if err != nil {
		return err
	}
	dsthashRaw, err := ei.N(patchmap).M("dsthash").Bytes()
	if err != nil {
		return err
	}
	data, err := ei.N(patchmap).M("data").Bytes()
	if err != nil {
		return err
	}

	var rdata []byte
	if p.rollbackType == "patch" && p.createRollback {
		rdata, err = ei.N(patchmap).M("rdata").Bytes()
		if err != nil {
			return fmt.Errorf("selected rollback type is \"patch\" but cannot get the rollback data data in patch file: %w", err)
		}
	}

	srchash := Sha256B64(srchashRaw)
	dsthash := Sha256B64(dsthashRaw)

	upgradeEnvPath, rollbackEnvPath, upgradeBinPath, rollbackBinPath, exitType := getRawParams(p)

	upgradeBinPathMsg := upgradeBinPath
	if !filepath.IsAbs(upgradeBinPathMsg) {
		upgradeBinPathMsg = fmt.Sprintf("(relative to idefix binary) %s", upgradeBinPathMsg)
	}
	rollbackBinPathMsg := rollbackBinPath
	if !filepath.IsAbs(rollbackBinPathMsg) {
		rollbackBinPathMsg = fmt.Sprintf("(relative to idefix binary) %s", rollbackBinPathMsg)
	}

	updateParamsTable := pterm.TableData{
		{"Upgrade params", ""},
		{"Target", "idefix"},
		{"Upgrade patch size", fmt.Sprintf("%d", len(data))},
		{"Source hash", srchash},
		{"Destination hash", dsthash},
		{"Create rollback", fmt.Sprintf("%v", p.createRollback)},
	}
	if p.createRollback {
		updateParamsTable = append(updateParamsTable, []string{"Rollback type", p.rollbackType})
	}
	updateParamsTable = append(updateParamsTable, pterm.TableData{
		{"Upgrade path", upgradeBinPathMsg},
		{"Rollback path", rollbackBinPathMsg},
		{"Stability time (s)", fmt.Sprintf("%v", p.stabilitySecs)},
		{"Healthy time (s)", fmt.Sprintf("%v", p.healthySecs)},
		{"Check ppp link", fmt.Sprintf("%v", p.checkPPP)},
		{"Check transport link", fmt.Sprintf("%v", p.checkTransport)},
		{"Stop countdown (s)", fmt.Sprintf("%v", p.stopToutSecs)},
		{"Halt timeout (s)", fmt.Sprintf("%v", p.haltToutSecs)},
	}...)

	pterm.DefaultTable.WithHasHeader().WithData(updateParamsTable).Render()
	fmt.Println()

	ic, err := getConnectedClient()
	if err != nil {
		return fmt.Errorf("cannot connect to the server: %w", err)
	}
	msg := map[string]interface{}{
		"report": false,
	}
	ret, err := ic.Call(p.address, &m.Message{To: "sys.cmd.info", Data: msg}, time.Second*10)
	if err != nil {
		return fmt.Errorf("cannot get device info: %w", err)
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

	// Check idefix file hash
	actualSrcHash, err := idefixgo.FileSHA256b64(ic, p.address, IdefixExecIdefixRelativePath, p.tout)
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
		{"Actual idefix hash", actualSrcHash},
	}).Render()

	fmt.Println()

	if actualSrcHash != srchash {
		return fmt.Errorf("idefix hash mismatch: actual %s != expected %s", actualSrcHash, srchash)
	}

	if result, _ := pterm.DefaultInteractiveConfirm.Show(); !result {
		return nil
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(false).Start("Sending update...")
	defer spinner.Stop()

	pterm.Success.Println("Patch completed! Device should reboot now...")
	return nil
}
