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

	cmdUpdateSendPatch.Flags().StringP("device", "d", "", "Device ID")
	cmdUpdateSendPatch.Flags().StringP("patch", "p", "", "Patch file")
	cmdUpdateSendPatch.MarkFlagRequired("device")
	cmdUpdateSendPatch.MarkFlagRequired("patch")
	cmdUpdate.AddCommand(cmdUpdateSendPatch)

	cmdUpdateSendFile.Flags().StringP("device", "d", "", "Device ID")
	cmdUpdateSendFile.Flags().StringP("file", "f", "", "Update file")
	cmdUpdateSendFile.MarkFlagRequired("device")
	cmdUpdateSendFile.MarkFlagRequired("file")
	cmdUpdate.AddCommand(cmdUpdateSendFile)

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

var cmdUpdateSendPatch = &cobra.Command{
	Use:   "send-patch",
	Short: "Send a patch to a remote device",
	RunE:  cmdUpdateSendPatchRunE,
}

var cmdUpdateSendFile = &cobra.Command{
	Use:   "send-file",
	Short: "Send file to update a remote device",
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
	src, err := cmd.Flags().GetString("source")
	if err != nil {
		return err
	}
	dst, err := cmd.Flags().GetString("destination")
	if err != nil {
		return err
	}

	patchbytes, srchash, dsthash, err := createPatch(src, dst)
	if err != nil {
		return err
	}

	patch := map[string]interface{}{
		"data":        patchbytes,
		"srchash:hex": srchash,
		"dsthash:hex": dsthash,
	}

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
		return fmt.Errorf("Source hash is %s, Expected: %s", hex.EncodeToString(psrchash[:]), hex.EncodeToString(srchash[:]))
	}

	newbytes, err := bspatch.Bytes(srcbytes, pdata)
	if err != nil {
		return err
	}

	dsthash := sha256.Sum256(newbytes)

	if !bytes.Equal(pdsthash, dsthash[:]) {
		return fmt.Errorf("Patched file hash is %s, Expected: %s", hex.EncodeToString(dsthash[:]), hex.EncodeToString(pdsthash[:]))
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

func cmdUpdateSendPatchRunE(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("device")
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

	ic, err := getConnectedClient()
	if err != nil {
		return fmt.Errorf("Cannot connect to the server: %w", err)
	}

	ret, err := ic.Call(addr, &idefixgo.Message{To: "sys.cmd.info"}, time.Second*10)
	if err != nil {
		return fmt.Errorf("Cannot get device info: %w", err)
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
	if result, _ := pterm.DefaultInteractiveConfirm.Show(); !result {
		return nil
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(false).Start("Sending update...")

	ret, err = ic.Call(addr, &idefixgo.Message{To: "updater.cmd.patch", Data: map[string]interface{}{"method": "bytes", "srchash": srchash, "dsthash": dsthash, "data": data}}, time.Hour*24)
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
	addr, err := cmd.Flags().GetString("device")
	if err != nil {
		return err
	}

	updatefile, err := cmd.Flags().GetString("file")
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
	if result, _ := pterm.DefaultInteractiveConfirm.Show(); !result {
		return nil
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(false).Start("Sending update...")

	ret, err = ic.Call(addr, &idefixgo.Message{To: "updater.cmd.update", Data: map[string]interface{}{"method": "bytes", "dsthash": dsthash, "data": updatebytes}}, time.Hour*24)
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
