package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"syscall"

	idefixgo "github.com/nayarsystems/idefix-go"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var defaultFileModeRaw uint32 = 0744

func init() {
	cmdOsFile.PersistentFlags().StringP("src", "s", "", "source path")

	cmdOs.AddCommand(cmdOsFile)

	cmdOsFileRead.PersistentFlags().StringP("dst", "d", "", "destination path")
	cmdOsFileRead.PersistentFlags().Uint32P("mode", "m", 0, fmt.Sprintf("file mode (default 0%o)", defaultFileModeRaw))
	cmdOsFileRead.MarkPersistentFlagRequired("src")
	cmdOsFileRead.MarkPersistentFlagRequired("dst")
	cmdOsFile.AddCommand(cmdOsFileRead)

	cmdOsFileWrite.PersistentFlags().StringP("dst", "d", "", "destination path")
	cmdOsFileWrite.PersistentFlags().Uint32P("mode", "m", 0, fmt.Sprintf("file mode (default 0%o)", defaultFileModeRaw))
	cmdOsFileWrite.MarkPersistentFlagRequired("src")
	cmdOsFileWrite.MarkPersistentFlagRequired("dst")
	cmdOsFile.AddCommand(cmdOsFileWrite)

	cmdOsFileMove.PersistentFlags().StringP("dst", "d", "", "destination path")
	cmdOsFileMove.PersistentFlags().Uint32P("mode", "m", 0, fmt.Sprintf("file mode (default 0%o)", defaultFileModeRaw))
	cmdOsFileMove.MarkPersistentFlagRequired("src")
	cmdOsFileMove.MarkPersistentFlagRequired("dst")
	cmdOsFile.AddCommand(cmdOsFileMove)

	cmdOsFileSha256.PersistentFlags().StringP("file", "f", "", "file to get SHA256 hash")
	cmdOsFileSha256.MarkPersistentFlagRequired("file")
	cmdOsFile.AddCommand(cmdOsFileSha256)
}

var cmdOsFile = &cobra.Command{
	Use:   "file",
	Short: "file related commands",
}

var cmdOsFileRead = &cobra.Command{
	Use:   "read",
	Short: "read file in remote device",
	RunE:  cmdOsFileReadRunE,
}

var cmdOsFileWrite = &cobra.Command{
	Use:   "write",
	Short: "write file in remote device",
	RunE:  cmdOsFileWriteRunE,
}

var cmdOsFileMove = &cobra.Command{
	Use:   "move",
	Short: "move file in remote device",
	RunE:  cmdOsFileMoveRunE,
}

var cmdOsFileSha256 = &cobra.Command{
	Use:   "sha256",
	Short: "get SHA256 hash of file in remote device",
	RunE:  cmdOsFileSha256RunE,
}

type fileBaseParams struct {
	osBaseParams
	srcPath  string
	dstPath  string
	fileMode os.FileMode
}

func getRWFileBaseParams(cmd *cobra.Command) (params fileBaseParams, err error) {
	params.osBaseParams, err = getOsBaseParams(cmd)
	if err != nil {
		return
	}

	params.srcPath, err = cmd.Flags().GetString("src")
	if err != nil {
		return
	}

	params.dstPath, err = cmd.Flags().GetString("dst")
	if err != nil {
		return
	}

	fileModeRaw, err := cmd.Flags().GetUint32("mode")
	if err != nil {
		return
	}
	if fileModeRaw == 0 {
		fileModeRaw = uint32(defaultFileModeRaw)
	}

	params.fileMode = os.FileMode(fileModeRaw)

	return
}

func cmdOsFileReadRunE(cmd *cobra.Command, args []string) (err error) {
	params, err := getRWFileBaseParams(cmd)
	if err != nil {
		return
	}

	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Device", ""},
		{"Address", params.address},
		{"Source file (device)", params.srcPath},
		{"Destination file (host)", params.dstPath},
		{"File mode", fmt.Sprintf("%o (%v)", params.fileMode, params.fileMode)},
	}).Render()

	fmt.Println()

	if result, _ := pterm.DefaultInteractiveConfirm.Show(); !result {
		return nil
	}

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}

	data, err := idefixgo.FileRead(ic, params.address, params.srcPath, params.timeout)
	if err != nil {
		return err
	}

	hash, err := idefixgo.FileSHA256Hex(ic, params.address, params.srcPath, params.timeout)
	if err != nil {
		return err
	}

	// Check hash of data
	dataHash := Sha256Hex(data)
	if dataHash != hash {
		return fmt.Errorf("read error. Hash mismatch: %s != %s", dataHash, hash)
	}

	err = os.WriteFile(params.dstPath, data, params.fileMode)
	if err != nil {
		return err
	}

	syscall.Sync()
	return
}

func cmdOsFileWriteRunE(cmd *cobra.Command, args []string) (err error) {
	params, err := getRWFileBaseParams(cmd)
	if err != nil {
		return
	}

	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Device", ""},
		{"Address", params.address},
		{"Source file (host)", params.srcPath},
		{"Destination file (device)", params.dstPath},
		{"File mode", fmt.Sprintf("%o (%v)", params.fileMode, params.fileMode)},
	}).Render()

	if result, _ := pterm.DefaultInteractiveConfirm.Show(); !result {
		return nil
	}

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}

	srcBytes, err := os.ReadFile(params.srcPath)
	if err != nil {
		return err
	}

	srcBytesHash := Sha256Hex(srcBytes)

	dstBytesHash, err := idefixgo.FileWrite(ic, params.address, params.dstPath, srcBytes, params.fileMode, params.timeout)
	if err != nil {
		return err
	}

	if srcBytesHash != dstBytesHash {
		return fmt.Errorf("write error. Resulting hash mismatch: %s != %s", srcBytesHash, dstBytesHash)
	}

	return
}

func cmdOsFileMoveRunE(cmd *cobra.Command, args []string) (err error) {
	params, err := getRWFileBaseParams(cmd)
	if err != nil {
		return
	}

	dstPath, err := cmd.Flags().GetString("dst")
	if err != nil {
		return
	}

	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Device", ""},
		{"Address", params.address},
		{"Source file (device)", params.srcPath},
		{"Destination file (device)", dstPath},
	}).Render()

	if result, _ := pterm.DefaultInteractiveConfirm.Show(); !result {
		return nil
	}

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}

	err = idefixgo.Move(ic, params.address, params.srcPath, dstPath, params.timeout)
	if err != nil {
		return err
	}

	return
}

func cmdOsFileSha256RunE(cmd *cobra.Command, args []string) (err error) {
	params, err := getOsBaseParams(cmd)
	if err != nil {
		return
	}

	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return
	}

	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Device", ""},
		{"Address", params.address},
		{"File", file},
	}).Render()

	if result, _ := pterm.DefaultInteractiveConfirm.Show(); !result {
		return nil
	}

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}

	hash, err := idefixgo.FileSHA256(ic, params.address, file, params.timeout)
	if err != nil {
		return err
	}
	// print hash in hex
	fmt.Printf("HEX: %x\n", hash)

	// print hash in b64
	hashB64 := base64.StdEncoding.EncodeToString(hash)
	fmt.Println("B64: " + hashB64)
	return
}
