package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabstv/go-bsdiff/pkg/bsdiff"
	"github.com/gabstv/go-bsdiff/pkg/bspatch"
	"github.com/jaracil/ei"
	idefixgo "github.com/nayarsystems/idefix-go"
	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/nayarsystems/idefix-go/normalize"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	cmdUpdate.PersistentFlags().StringP("target", "t", "idefix", "Target: launcher,idefix. Default: idefix")

	cmdUpdateCreate.Flags().StringP("source", "s", "", "Source")
	cmdUpdateCreate.Flags().StringP("destination", "d", "", "Destination")
	cmdUpdateCreate.Flags().StringP("output", "o", "", "Output")
	cmdUpdateCreate.Flags().BoolP("rollback", "r", false, "Also include a patch for rollback")
	cmdUpdateCreate.MarkFlagRequired("source")
	cmdUpdateCreate.MarkFlagRequired("destination")
	cmdUpdate.AddCommand(cmdUpdateCreate)

	cmdUpdateApply.Flags().StringP("source", "s", "", "Source file")
	cmdUpdateApply.Flags().StringP("patch", "p", "", "Patch file")
	cmdUpdateApply.Flags().StringP("output", "o", "", "Output file")
	cmdUpdateApply.Flags().BoolP("inplace", "i", false, "Patch and overwrite source file")
	cmdUpdateApply.MarkFlagsMutuallyExclusive("output", "inplace")
	cmdUpdateApply.MarkFlagRequired("file")
	cmdUpdateApply.MarkFlagRequired("patch")
	cmdUpdate.AddCommand(cmdUpdateApply)

	cmdUpdateSendPatch.Flags().StringP("patch", "p", "", "Patch file")
	cmdUpdateSendPatch.MarkFlagRequired("patch")
	cmdUpdateSend.AddCommand(cmdUpdateSendPatch)

	cmdUpdateSendFile.Flags().StringP("file", "f", "", "Update file")
	cmdUpdateSendFile.MarkFlagRequired("file")
	cmdUpdateSendFile.Flags().BoolP("rollback", "r", false, "Also request device to save a rollback file (full .bin)")
	cmdUpdateSend.AddCommand(cmdUpdateSendFile)

	cmdUpdateSend.PersistentFlags().StringP("address", "a", "", "Device address")
	cmdUpdateSend.PersistentFlags().Uint("stability-secs", 60, "Indicates the duration of the test execution in seconds")
	cmdUpdateSend.PersistentFlags().Uint("healthy-secs", 10, "Only used if at least one check is enabled. Indicates the minimum number of seconds positively validating the checks")
	cmdUpdateSend.PersistentFlags().Bool("check-ppp", false, "Check ppp link after upgrade")
	cmdUpdateSend.PersistentFlags().Bool("check-tr", false, "Check transport link after upgrade")
	cmdUpdateSend.PersistentFlags().Uint("stop-countdown", 10, "Stop countdown before stopping idefix in seconds")
	cmdUpdateSend.PersistentFlags().Uint("halt-timeout", 10, "Halt timeout in seconds")
	cmdUpdateSend.PersistentFlags().StringP("upgrade-path", "", "", "alternative upgrade file's path (absolute or relative to idefix binary)")
	cmdUpdateSend.PersistentFlags().StringP("rollback-path", "", "", "alternative rollback file's path (absolute or relative to idefix binary)")
	cmdUpdate.AddCommand(cmdUpdateSend)

	rootCmd.AddCommand(cmdUpdate)
}

var cmdUpdate = &cobra.Command{
	Use:   "update",
	Short: "Create, apply and send binary updates",
}

var cmdUpdateCreate = &cobra.Command{
	Use:   "create",
	Short: "Generate a patch file for updates",
	RunE:  cmdUpdateCreateRunE,
}

var cmdUpdateApply = &cobra.Command{
	Use:   "apply",
	Short: "Apply a patch to a local file",
	RunE:  cmdUpdateApplyRunE,
}

var cmdUpdateSend = &cobra.Command{
	Use:   "send",
	Short: "Send an update to a remote device",
}

var cmdUpdateSendPatch = &cobra.Command{
	Use:   "patch",
	Short: "Send a patch",
	RunE:  cmdUpdateSendPatchRunE,
}

var cmdUpdateSendFile = &cobra.Command{
	Use:   "file",
	Short: "Send a file",
	RunE:  cmdUpdateSendFileRunE,
}

const (
	IdefixExecIdefixRelativePath = "./idefix"
	IdefixUpgradeBasename        = "idefix_upgrade"
	IdefixRollbackBasename       = "idefix_rollback"
	LauncherUpgradeBasename      = "launcher_upgrade"
)

const (
	// Indicates the duration of the test execution in seconds
	ENV_IDEFIX_STABILITY_SECS = "ENV_IDEFIX_STABILITY_SECS"
	// Indicates the minimum number of seconds positively validating the checks.
	// So that, this parameter is only used if at least one of the healthy checks is enabled.
	// Note that this value must be less than the duration of the test execution (ENV_IDEFIX_STABILITY_SECS).
	ENV_IDEFIX_HEALTHY_SECS = "ENV_IDEFIX_HEALTHY_SECS"
	// Enables check of ppp link
	ENV_IDEFIX_PPP_CHECK = "ENV_IDEFIX_PPP_CHECK"
	// Enables check of transport link
	ENV_IDEFIX_TRANSPORT_CHECK = "ENV_IDEFIX_TRANSPORT_CHECK"
)
const (
	PARAM_UPDATE_FILE_PATH = "PARAM_UPDATE_FILE_PATH"
	PARAM_UPDATE_SRC_HASH  = "PARAM_UPDATE_SRC_HASH"
	PARAM_UPDATE_DST_HASH  = "PARAM_UPDATE_DST_HASH"
)

func createPatch(oldpath string, newpath string) ([]byte, string, string, error) {
	srcbytes, err := os.ReadFile(oldpath)
	if err != nil {
		return []byte{}, "", "", err
	}

	srchash := sha256.Sum256(srcbytes)

	dstbytes, err := os.ReadFile(newpath)
	if err != nil {
		return []byte{}, "", "", err
	}
	dsthash := sha256.Sum256(dstbytes)

	d, err := bsdiff.Bytes(srcbytes, dstbytes)
	if err != nil {
		return []byte{}, "", "", err
	}

	return d, hex.EncodeToString(srchash[:]), hex.EncodeToString(dsthash[:]), nil
}

func cmdUpdateCreateRunE(cmd *cobra.Command, args []string) error {
	createRollbackPatch, err := cmd.Flags().GetBool("rollback")
	if err != nil {
		return err
	}
	src, err := cmd.Flags().GetString("source")
	if err != nil {
		return err
	}
	dst, err := cmd.Flags().GetString("destination")
	if err != nil {
		return err
	}
	spinner, _ := pterm.DefaultSpinner.WithShowTimer(false).Start("Creating patch...")

	patchbytes, srchash, dsthash, err := createPatch(src, dst)
	if err != nil {
		return err
	}

	patch := map[string]interface{}{
		"data":        patchbytes,
		"srchash:hex": srchash,
		"dsthash:hex": dsthash,
	}
	patchhash := sha256.Sum256(patchbytes)
	patchhashStr := hex.EncodeToString(patchhash[:])
	tp := pterm.TableData{
		{"Patch", ""},
		{"Src hash", srchash},
		{"Dst hash", dsthash},
		{"Patch hash", patchhashStr},
	}
	if createRollbackPatch {
		rpatchbytes, _, _, err := createPatch(dst, src)
		if err != nil {
			return err
		}
		rpatchbyteshash := sha256.Sum256(rpatchbytes)
		rpatchbyteshashStr := hex.EncodeToString(rpatchbyteshash[:])
		patch["rdata"] = rpatchbytes
		tp = append(tp, []string{"Rollback patch hash", rpatchbyteshashStr})
	}
	spinner.Stop()
	pterm.DefaultTable.WithHasHeader().WithData(tp).Render()

	normalize.EncodeTypes(patch, &normalize.EncodeTypesOpts{BytesToB64: true, Compress: true})

	j, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	if !cmd.Flags().Changed("output") {
		fmt.Printf("%s\n", string(j))
	} else {
		out, err := cmd.Flags().GetString("output")
		if err != nil {
			return err
		}

		ioutil.WriteFile(out, j, 0644)
	}
	pterm.Success.Println("Patch created!")

	return nil
}

func cmdUpdateApplyRunE(cmd *cobra.Command, args []string) error {
	src, err := cmd.Flags().GetString("source")
	if err != nil {
		return err
	}
	patch, err := cmd.Flags().GetString("patch")
	if err != nil {
		return err
	}

	srcbytes, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	srchash := sha256.Sum256(srcbytes)

	patchbytes, err := os.ReadFile(patch)
	if err != nil {
		return err
	}

	patchmap := make(map[string]interface{})

	if err := json.Unmarshal(patchbytes, &patchmap); err != nil {
		return err
	}

	if err := normalize.DecodeTypes(patchmap); err != nil {
		return err
	}

	psrchash, err := ei.N(patchmap).M("srchash").Bytes()
	if err != nil {
		return err
	}

	pdsthash, err := ei.N(patchmap).M("dsthash").Bytes()
	if err != nil {
		return err
	}

	pdata, err := ei.N(patchmap).M("data").Bytes()
	if err != nil {
		return err
	}

	if !bytes.Equal(srchash[:], psrchash) {
		return fmt.Errorf("source hash is %s, Expected: %s", hex.EncodeToString(psrchash[:]), hex.EncodeToString(srchash[:]))
	}

	newbytes, err := bspatch.Bytes(srcbytes, pdata)
	if err != nil {
		return err
	}

	dsthash := sha256.Sum256(newbytes)

	if !bytes.Equal(pdsthash, dsthash[:]) {
		return fmt.Errorf("patched file hash is %s, Expected: %s", hex.EncodeToString(dsthash[:]), hex.EncodeToString(pdsthash[:]))
	}

	if !cmd.Flags().Changed("output") && !cmd.Flags().Changed("inplace") {
		fmt.Println("The patch can be applied (no files were modified)")
		return nil
	}

	outpath := ""

	if b, err := cmd.Flags().GetBool("inplace"); b && err == nil {
		outpath = src
	} else {
		outpath, err = cmd.Flags().GetString("output")
		if err != nil {
			return err
		}
	}

	stat, err := os.Stat(src)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(outpath, newbytes, stat.Mode())
}

func remoteExec(ic *idefixgo.Client, addr string, cmd string, timeout time.Duration) (interface{}, error) {
	m, err := ic.Call(addr, &m.Message{To: "os.cmd.exec", Data: map[string]interface{}{
		"command": cmd,
	}}, timeout)

	if err != nil {
		return nil, err
	}

	if !ei.N(m.Data).M("success").BoolZ() {
		return nil, fmt.Errorf("command failed")
	}

	return m.Data, err
}

type updateParams struct {
	target                  m.TargetExec
	targetStr               string
	checkPPP                bool
	checkTransport          bool
	healthySecs             uint
	stabilitySecs           uint
	stopToutSecs            uint
	haltToutSecs            uint
	createRollback          bool
	upgradeType             string
	rollbackType            string
	alternativeUpgradePath  string
	alternativeRollbackPath string
}

func getUpdateParams(cmd *cobra.Command) (p *updateParams, err error) {
	p = &updateParams{}
	p.targetStr, err = cmd.Flags().GetString("target")
	if err != nil {
		return
	}
	switch p.targetStr {
	case "launcher":
		p.target = m.LauncherTargetExec
	case "idefix":
		p.target = m.IdefixTargetExec
	default:
		return nil, fmt.Errorf("invalid target")
	}

	p.checkPPP, err = cmd.Flags().GetBool("check-ppp")
	if err != nil {
		return
	}

	p.checkTransport, err = cmd.Flags().GetBool("check-tr")
	if err != nil {
		return
	}

	p.healthySecs, err = cmd.Flags().GetUint("healthy-secs")
	if err != nil {
		return
	}

	p.stabilitySecs, err = cmd.Flags().GetUint("stability-secs")
	if err != nil {
		return
	}

	p.stopToutSecs, err = cmd.Flags().GetUint("stop-countdown")
	if err != nil {
		return
	}

	p.haltToutSecs, err = cmd.Flags().GetUint("halt-timeout")
	if err != nil {
		return
	}
	p.createRollback, err = cmd.Flags().GetBool("rollback")
	if err != nil {
		return
	}
	p.alternativeUpgradePath, err = cmd.Flags().GetString("upgrade-path")
	if err != nil {
		return
	}
	if p.alternativeUpgradePath != "" && !strings.HasSuffix(p.alternativeUpgradePath, ".bin") && !strings.HasSuffix(p.alternativeUpgradePath, ".patch") {
		return nil, fmt.Errorf("invalid upgrade extension (.bin|.patch )")
	}
	p.alternativeRollbackPath, err = cmd.Flags().GetString("rollback-path")
	if err != nil {
		return
	}
	if p.alternativeRollbackPath != "" && !strings.HasSuffix(p.alternativeRollbackPath, ".bin") && !strings.HasSuffix(p.alternativeRollbackPath, ".patch") {
		return nil, fmt.Errorf("invalid rollback extension (.bin|.patch )")
	}

	if p.target == m.LauncherTargetExec && p.createRollback {
		return nil, fmt.Errorf("rollback not supported for launcher")
	}
	return
}

func cmdUpdateSendPatchRunE(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
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

	srchash, err := ei.N(patchmap).M("srchash").Bytes()
	if err != nil {
		return err
	}
	dsthash, err := ei.N(patchmap).M("dsthash").Bytes()
	if err != nil {
		return err
	}
	data, err := ei.N(patchmap).M("data").Bytes()
	if err != nil {
		return err
	}

	var hasRollback bool
	rdata, err := ei.N(patchmap).M("rdata").Bytes()
	if err == nil {
		hasRollback = true
	}

	p, err := getUpdateParams(cmd)
	if err != nil {
		return err
	}

	ic, err := getConnectedClient()
	if err != nil {
		return fmt.Errorf("cannot connect to the server: %w", err)
	}
	msg := map[string]interface{}{
		"report": false,
	}
	ret, err := ic.Call(addr, &m.Message{To: "sys.cmd.info", Data: msg}, time.Second*10)
	if err != nil {
		return fmt.Errorf("cannot get device info: %w", err)
	}
	address, version, bootcnt, err := getDevInfoFromMsg(ret.Data)
	if err != nil {
		return err
	}

	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Patch", "", ""},
		{"Size", fmt.Sprintf("%d", len(data))},
		{"Source Hash", hex.EncodeToString(srchash)},
		{"Destination Hash", hex.EncodeToString(dsthash)},
	}).Render()

	fmt.Println()

	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Device", ""},
		{"Address", address},
		{"Boot Counter", fmt.Sprintf("%d", bootcnt)},
		{"Version", version},
	}).Render()

	fmt.Println()

	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Update params", "", ""},
		{"Stability time (s)", fmt.Sprintf("%v", p.stabilitySecs)},
		{"Healthy time (s)", fmt.Sprintf("%v", p.healthySecs)},
		{"Check ppp link", fmt.Sprintf("%v", p.checkPPP)},
		{"Check transport link", fmt.Sprintf("%v", p.checkTransport)},
		{"Stop countdown (s)", fmt.Sprintf("%v", p.stopToutSecs)},
		{"Halt timeout (s)", fmt.Sprintf("%v", p.haltToutSecs)},
	}).Render()

	fmt.Println()

	if result, _ := pterm.DefaultInteractiveConfirm.Show(); !result {
		return nil
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(false).Start("Sending update...")

	patchMsg := map[string]interface{}{
		"method":         "bytes",
		"target":         p.target,
		"srchash":        srchash,
		"dsthash":        dsthash,
		"data":           data,
		"check_ppp":      p.checkPPP,
		"check_tr":       p.checkTransport,
		"stability_secs": p.stabilitySecs,
		"healthy_secs":   p.healthySecs,
		"stop_countdown": p.stopToutSecs,
		"halt_timeout":   p.haltToutSecs,
	}
	if hasRollback {
		patchMsg["rdata"] = rdata
	}
	ret, err = ic.Call(addr, &m.Message{To: "updater.cmd.patch", Data: patchMsg}, time.Hour*24)
	spinner.Stop()

	if err != nil {
		return err
	}

	if ret.Err != "" {
		return fmt.Errorf(ret.Err)
	}

	pterm.Success.Println("Patch completed! Device should reboot now...")
	return nil
}

func cmdUpdateSendFileRunE(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
	}

	updatefile, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	toutMs, err := cmd.Flags().GetUint("timeout")
	if err != nil {
		return err
	}
	tout := time.Duration(toutMs) * time.Millisecond

	p, err := getUpdateParams(cmd)
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
		{"Update params", ""},
		{"Target", p.targetStr},
		{"Update file size", KB(uint64(len(updatebytes)))},
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
	ret, err := ic.Call(addr, &m.Message{To: "sys.cmd.info", Data: msg}, time.Second)
	if err != nil {
		return err
	}
	address, version, bootcnt, err := getDevInfoFromMsg(ret.Data)
	if err != nil {
		return err
	}

	// Check space available
	freeSpace, err := idefixgo.GetFree(ic, addr, "", tout)
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

	if freeSpace < uint64(len(updatebytes)) {
		return fmt.Errorf("not enough space available: update (%d) > free (%d)", len(updatebytes), freeSpace)
	}

	fmt.Println()

	if result, _ := pterm.DefaultInteractiveConfirm.Show(); !result {
		return nil
	}
	spinner, _ := pterm.DefaultSpinner.WithShowTimer(false).Start("")
	defer spinner.Stop()

	spinner.UpdateText("sending upgrade env file...")
	err = sendEnvFile(ic, addr, p, upgradeEnvPath, false, "", "", tout)
	if err != nil {
		return err
	}

	spinner.UpdateText("sending upgrade bin file...")
	receivedHash, err := idefixgo.FileWrite(ic, addr, upgradeBinPath, updatebytes, 0744, tout)
	if err != nil {
		return err
	}
	if receivedHash != dsthash {
		return fmt.Errorf("received upgrade hash (%s) != expected upgrade hash (%s)", receivedHash, dsthash)
	}

	if p.createRollback && p.target == m.IdefixTargetExec {
		spinner.UpdateText("sending rollback env file...")
		err = sendEnvFile(ic, addr, p, rollbackEnvPath, true, "", "", tout)
		if err != nil {
			return err
		}
		spinner.UpdateText("backup idefix file...")
		err = idefixgo.FileCopy(ic, addr, IdefixExecIdefixRelativePath, rollbackBinPath, tout)
		if err != nil {
			return err
		}
	}

	spinner.UpdateText("sending update request...")
	res, err := idefixgo.ExitToUpdate(ic, addr,
		exitType,
		"",
		time.Duration(p.stopToutSecs)*time.Second,
		time.Duration(p.haltToutSecs)*time.Second,
		tout)
	if err != nil {
		return err
	}
	pterm.Success.Printf("Response. %#v\nUpdate completed!\nDevice should reboot now...\n", res)
	return nil
}

func getDevInfoFromMsg(data interface{}) (address, version string, bootCnt int, err error) {
	var devInfo map[string]interface{}
	devInfo, err = ei.N(data).M("devInfo").MapStr()
	if err != nil {
		return
	}
	address, err = ei.N(devInfo).M("address").String()
	if err != nil {
		return
	}
	bootCnt, err = ei.N(devInfo).M("bootCnt").Int()
	if err != nil {
		return
	}
	version, err = ei.N(devInfo).M("version").String()
	if err != nil {
		return
	}
	return
}

func sendEnvFile(ic *idefixgo.Client, addr string, p *updateParams, envFilePath string, isRollback bool, srcHash, dstHash string, tout time.Duration) (err error) {
	// Build and send upgrade's env file (if not empty)
	envFileData, err := buildEnvFile(p, isRollback, srcHash, dstHash)
	if err != nil {
		return err
	}
	envFileHash := Sha256B64(envFileData)
	receivedHash, err := idefixgo.FileWrite(ic, addr, envFilePath, envFileData, 0744, tout)
	if err != nil {
		return err
	}
	if receivedHash != envFileHash {
		return fmt.Errorf("received env hash (%s) != expected env hash (%s)", receivedHash, envFileHash)
	}
	return
}

func buildEnvFile(p *updateParams, isRollback bool, srcHash, dstHash string) ([]byte, error) {
	var buf bytes.Buffer
	if p.target == m.IdefixTargetExec && !isRollback {
		buf.WriteString(fmt.Sprintf("%s=%d\n", ENV_IDEFIX_STABILITY_SECS, p.stabilitySecs))
		buf.WriteString(fmt.Sprintf("%s=%d\n", ENV_IDEFIX_HEALTHY_SECS, p.healthySecs))
		buf.WriteString(fmt.Sprintf("%s=%v\n", ENV_IDEFIX_PPP_CHECK, p.checkPPP))
		buf.WriteString(fmt.Sprintf("%s=%v\n", ENV_IDEFIX_TRANSPORT_CHECK, p.checkTransport))
	}
	if srcHash != "" {
		buf.WriteString(fmt.Sprintf("%s=%v\n", PARAM_UPDATE_SRC_HASH, srcHash))
	}
	if dstHash != "" {
		buf.WriteString(fmt.Sprintf("%s=%v\n", PARAM_UPDATE_DST_HASH, dstHash))
	}
	if isRollback && p.alternativeRollbackPath != "" {
		buf.WriteString(fmt.Sprintf("%s=%v\n", PARAM_UPDATE_FILE_PATH, getLauncherRelativePath(p.alternativeRollbackPath)))
	}
	if !isRollback && p.alternativeUpgradePath != "" {
		buf.WriteString(fmt.Sprintf("%s=%v\n", PARAM_UPDATE_FILE_PATH, getLauncherRelativePath(p.alternativeUpgradePath)))
	}
	return buf.Bytes(), nil
}

// get either an absolute or launcher relative path from either an absolute or idefix relative path
func getLauncherRelativePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join("./idefix/", path)
}

func getRawParams(p *updateParams) (upgradeEnvPath, rollbackEnvPath, upgradeBinPath, rollbackBinPath string, exitType int) {
	if p.target == m.IdefixTargetExec {
		// IDEFIX
		exitType = m.UpdateTypeIdefixUpgrade
		upgradeEnvPath = filepath.Join("../updates", fmt.Sprintf("%s.env", IdefixUpgradeBasename))
		upgradeBinPath = filepath.Join("../updates", fmt.Sprintf("%s.%s", IdefixUpgradeBasename, p.upgradeType))
		rollbackEnvPath = filepath.Join("../updates", fmt.Sprintf("%s.env", IdefixRollbackBasename))
		rollbackBinPath = filepath.Join("../updates", fmt.Sprintf("%s.%s", IdefixRollbackBasename, p.rollbackType))
	} else {
		// LAUNCHER
		exitType = m.UpdateTypeLauncherUpgrade
		upgradeEnvPath = filepath.Join("../updates", fmt.Sprintf("%s.env", LauncherUpgradeBasename))
		upgradeBinPath = filepath.Join("../updates", fmt.Sprintf("%s.%s", LauncherUpgradeBasename, p.upgradeType))
		// rollback not supported
	}

	if p.alternativeUpgradePath != "" {
		upgradeBinPath = p.alternativeUpgradePath
	}
	if p.alternativeRollbackPath != "" {
		rollbackBinPath = p.alternativeRollbackPath
	}
	return
}

func KB(bytes uint64) string {
	return fmt.Sprintf("%.2f KB", float64(bytes)/math.Pow(2, 10))
}

func Sha256B64(bytes []byte) string {
	hash := sha256.Sum256(bytes)
	return base64.StdEncoding.EncodeToString(hash[:])
}
