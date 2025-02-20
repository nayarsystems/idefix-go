package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jaracil/ei"
	idefixgo "github.com/nayarsystems/idefix-go"
	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/spf13/cobra"
)

func init() {
	cmdUpdate.PersistentFlags().StringP("target", "t", "idefix", "Target: launcher,idefix. Default: idefix")

	cmdUpdateCreate.Flags().StringP("source", "s", "", "Source")
	cmdUpdateCreate.Flags().StringP("destination", "d", "", "Destination")
	cmdUpdateCreate.Flags().StringP("output", "o", "", "Output")
	cmdUpdateCreate.Flags().BoolP("rollback", "r", false, "Also include a rollback patch")
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
	cmdUpdateSendPatch.Flags().String("rollback-type", "bin", "Type of rollback file: bin (request backup file),patch (send rollback patch)")
	cmdUpdateSend.AddCommand(cmdUpdateSendPatch)

	cmdUpdateSendFile.Flags().StringP("file", "f", "", "Update file")
	cmdUpdateSendFile.MarkFlagRequired("file")
	cmdUpdateSend.AddCommand(cmdUpdateSendFile)

	cmdUpdateSend.PersistentFlags().StringP("address", "a", "", "Device address")
	cmdUpdateSend.PersistentFlags().String("reason", "", "Optinal reason for update")
	cmdUpdateSend.PersistentFlags().Uint("stability-secs", 300, "Indicates the duration of the test execution in seconds")
	cmdUpdateSend.PersistentFlags().Uint("healthy-secs", 60, "Only used if at least one check is enabled. Indicates the minimum number of seconds positively validating the checks")
	cmdUpdateSend.PersistentFlags().Bool("check-ppp", true, "Check ppp link after upgrade")
	cmdUpdateSend.PersistentFlags().Bool("check-tr", true, "Check transport link after upgrade")
	cmdUpdateSend.PersistentFlags().Uint("stop-countdown", 10, "Stop countdown before stopping idefix in seconds")
	cmdUpdateSend.PersistentFlags().Uint("halt-timeout", 10, "Halt timeout in seconds")
	cmdUpdateSend.PersistentFlags().StringP("upgrade-path", "", "", "Alternative upgrade file's path (absolute or relative to idefix binary)")
	cmdUpdateSend.PersistentFlags().BoolP("no-rollback", "", false, "Do not send/request a rollback file")
	cmdUpdateSend.PersistentFlags().StringP("rollback-path", "", "", "Alternative rollback file's path (absolute or relative to idefix binary)")
	cmdUpdateSend.PersistentFlags().UintP("timeout", "", 120000, "timeout in milliseconds")
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

type updateParams struct {
	address                 string
	reason                  string
	tout                    time.Duration
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

	toutMs, err := cmd.Flags().GetUint("timeout")
	if err != nil {
		return
	}
	p.tout = time.Duration(toutMs) * time.Millisecond

	p.address, err = cmd.Flags().GetString("address")
	if err != nil {
		return
	}

	p.reason, err = cmd.Flags().GetString("reason")
	if err != nil {
		return
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

	noRollback, err := cmd.Flags().GetBool("no-rollback")
	if err != nil {
		return
	}
	p.createRollback = !noRollback

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
	envFileHash := Sha256Hex(envFileData)
	receivedHash, err := idefixgo.FileWrite(ic, addr, envFilePath, envFileData, 0744, tout)
	if err != nil {
		return err
	}
	if receivedHash != envFileHash {
		return fmt.Errorf("received env hash (%s) != expected env hash (%s)", receivedHash, envFileHash)
	}
	return
}

func hexToB64(hexdata string) (b64data string, err error) {
	raw, err := hex.DecodeString(hexdata)
	if err != nil {
		return
	}
	b64data = base64.StdEncoding.EncodeToString(raw)
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
		srcHashB64, err := hexToB64(srcHash)
		if err != nil {
			return nil, err
		}
		buf.WriteString(fmt.Sprintf("%s=%v\n", PARAM_UPDATE_SRC_HASH, srcHashB64))
	}
	if dstHash != "" {
		dstHashB64, err := hexToB64(dstHash)
		if err != nil {
			return nil, err
		}
		buf.WriteString(fmt.Sprintf("%s=%v\n", PARAM_UPDATE_DST_HASH, dstHashB64))
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

func Sha256Hex(bytes []byte) string {
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}

func storeFileBackup(fileContent []byte) error {
	ucd, err := os.UserCacheDir()
	if err != nil {
		ucd = "$HOME"
	}

	dstFolder := filepath.Join(ucd, "idefix", "updates")
	if err := os.MkdirAll(dstFolder, 0755); err != nil {
		return err
	}

	hash := sha256.Sum256(fileContent)
	hashStr := hex.EncodeToString(hash[:])

	backupPath := filepath.Join(dstFolder, hashStr)

	if _, err := os.Stat(backupPath); err == nil {
		return nil
	}

	return os.WriteFile(backupPath, fileContent, 0644)
}
