package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/gabstv/go-bsdiff/pkg/bsdiff"
	"github.com/gabstv/go-bsdiff/pkg/bspatch"
	"github.com/jaracil/ei"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	idefixgo "gitlab.com/garagemakers/idefix-go"
	"gitlab.com/garagemakers/idefix/core/idefix/normalize"
)

func init() {
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
	cmdUpdateSend.AddCommand(cmdUpdateSendFile)

	cmdUpdateSend.PersistentFlags().StringP("address", "a", "", "Device address")
	cmdUpdateSend.PersistentFlags().Uint("stability-secs", 60, "Indicates the duration of the test execution in seconds")
	cmdUpdateSend.PersistentFlags().Uint("healthy-secs", 10, "Only used if at least one check is enabled. Indicates the minimum number of seconds positively validating the checks")
	cmdUpdateSend.PersistentFlags().Bool("check-ppp", false, "Check ppp link after upgrade")
	cmdUpdateSend.PersistentFlags().Bool("check-tr", false, "Check transport link after upgrade")
	cmdUpdateSend.PersistentFlags().Uint("stop-countdown", 10, "Stop countdown before stopping idefix in seconds")
	cmdUpdateSend.PersistentFlags().Uint("halt-timeout", 10, "Halt timeout in seconds")
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
	m, err := ic.Call(addr, &idefixgo.Message{To: "os.cmd.exec", Data: map[string]interface{}{
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
	checkPPP       bool
	checkTransport bool
	healthySecs    uint
	stabilitySecs  uint
	stopToutSecs   uint
	haltToutSecs   uint
}

func getUpdateParams(cmd *cobra.Command) (p *updateParams, err error) {
	p = &updateParams{}
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

	ret, err := ic.Call(addr, &idefixgo.Message{To: "sys.cmd.info"}, time.Second*10)
	if err != nil {
		return fmt.Errorf("cannot get device info: %w", err)
	}
	address, err := ei.N(ret.Data).M("Address").String()
	if err != nil {
		return err
	}
	bootcnt, err := ei.N(ret.Data).M("BootCnt").Int()
	if err != nil {
		return err
	}
	version, err := ei.N(ret.Data).M("Version").String()
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
	ret, err = ic.Call(addr, &idefixgo.Message{To: "updater.cmd.patch", Data: patchMsg}, time.Hour*24)
	spinner.Stop()

	if err != nil {
		return err
	}

	if ret.Err != nil {
		return ret.Err
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

	p, err := getUpdateParams(cmd)
	if err != nil {
		return err
	}

	updatebytes, err := os.ReadFile(updatefile)
	if err != nil {
		return err
	}

	dsthash := sha256.Sum256(updatebytes)
	if err != nil {
		return err
	}

	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Update", "", ""},
		{"Size", fmt.Sprintf("%d", len(updatebytes))},
		{"Destination Hash", hex.EncodeToString(dsthash[:])},
	}).Render()

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}

	ret, err := ic.Call(addr, &idefixgo.Message{To: "sys.cmd.info"}, time.Second)
	if err != nil {
		return err
	}
	address, err := ei.N(ret.Data).M("Address").String()
	if err != nil {
		return err
	}
	bootcnt, err := ei.N(ret.Data).M("BootCnt").Int()
	if err != nil {
		return err
	}
	version, err := ei.N(ret.Data).M("Version").String()
	if err != nil {
		return err
	}

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

	msg := map[string]interface{}{
		"method":         "bytes",
		"dsthash":        dsthash,
		"data":           updatebytes,
		"check_ppp":      p.checkPPP,
		"check_tr":       p.checkTransport,
		"stability_secs": p.stabilitySecs,
		"healthy_secs":   p.healthySecs,
		"stop_countdown": p.stopToutSecs,
		"halt_timeout":   p.haltToutSecs,
	}
	ret, err = ic.Call(addr, &idefixgo.Message{To: "updater.cmd.update", Data: msg}, time.Hour*24)
	spinner.Stop()

	if err != nil {
		return err
	}

	if ret.Err != nil {
		return ret.Err
	}

	pterm.Success.Println("Update completed! Device should reboot now...")
	return nil
}
