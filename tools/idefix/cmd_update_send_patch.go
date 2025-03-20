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
	upgradePatch, err := ei.N(patchmap).M("data").Bytes()
	if err != nil {
		return err
	}

	var rollbackPatch []byte
	if p.rollbackType == "patch" && p.createRollback {
		rollbackPatch, err = ei.N(patchmap).M("rdata").Bytes()
		if err != nil {
			return fmt.Errorf("selected rollback type is \"patch\" but cannot get the rollback data data in patch file: %w", err)
		}
	}

	srchash := hex.EncodeToString(srchashRaw)
	dsthash := hex.EncodeToString(dsthashRaw)

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
		{"Reason", p.reason},
		{"Upgrade patch size", KB(uint64(len(upgradePatch)))},
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

	ic, err := getConnectedClient()
	if err != nil {
		return fmt.Errorf("cannot connect to the server: %w", err)
	}
	msg := map[string]interface{}{
		"report": false,
	}
	ret, err := ic.Call(p.address, &m.Message{To: "sys.cmd.info", Data: msg}, p.tout)
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
	actualSrcHash, err := idefixgo.FileSHA256Hex(ic, p.address, IdefixExecIdefixRelativePath, p.tout)
	if err != nil {
		return err
	}

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

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(false).Start("sending upgrade env file...")
	defer spinner.Stop()

	err = sendEnvFile(ic, p.address, p, upgradeEnvPath, false, srchash, dsthash, p.tout)
	if err != nil {
		return err
	}

	patchHash := Sha256Hex(upgradePatch)
	spinner.UpdateText("sending upgrade patch file...")
	receivedHash, err := idefixgo.FileWriteInChunks(ic, p.address, upgradeBinPath, upgradePatch, 1024*256, 0744, p.tout)
	if err != nil {
		return err
	}
	if receivedHash != patchHash {
		return fmt.Errorf("received upgrade hash (%s) != expected upgrade hash (%s)", receivedHash, patchHash)
	}

	if p.createRollback {
		spinner.UpdateText("sending rollback env file...")
		err = sendEnvFile(ic, p.address, p, rollbackEnvPath, true, dsthash, srchash, p.tout)
		if err != nil {
			return err
		}
		if p.rollbackType == "patch" {
			patchHash := Sha256Hex(rollbackPatch)
			spinner.UpdateText("sending rollback patch file...")
			receivedHash, err := idefixgo.FileWriteInChunks(ic, p.address, rollbackBinPath, rollbackPatch, 1024*256, 0744, p.tout)
			if err != nil {
				return err
			}
			if receivedHash != patchHash {
				return fmt.Errorf("received upgrade hash (%s) != expected upgrade hash (%s)", receivedHash, patchHash)
			}

		} else {
			spinner.UpdateText("backup idefix file...")
			err = idefixgo.FileCopy(ic, p.address, IdefixExecIdefixRelativePath, rollbackBinPath, p.tout)
			if err != nil {
				return err
			}
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
