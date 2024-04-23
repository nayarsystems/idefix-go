package main

import (
	"fmt"
	"os"
	"syscall"
	"time"

	idefixgo "github.com/nayarsystems/idefix-go"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var defaultFileModeRaw uint32 = 0744

func init() {
	cmdOsFile.PersistentFlags().StringP("src", "s", "", "source path")
	cmdOsFile.PersistentFlags().Uint32P("mode", "m", 0, fmt.Sprintf("file mode (default 0%o)", defaultFileModeRaw))
	cmdOsFile.MarkPersistentFlagRequired("src")

	cmdOs.AddCommand(cmdOsFile)

	cmdOsFileRead.Flags().StringP("dst", "d", "", "destination path in host")
	cmdOsFileRead.MarkFlagRequired("dst")

	cmdOsFile.AddCommand(cmdOsFileRead)

	cmdOsFileWrite.Flags().StringP("dst", "d", "", "destination path in device")
	cmdOsFileWrite.MarkFlagRequired("dst")

	cmdOsFile.AddCommand(cmdOsFileWrite)
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

type fileBaseParams struct {
	address  string
	srcPath  string
	fileMode os.FileMode
	timeout  time.Duration
}

func getFileBaseParams(cmd *cobra.Command) (params fileBaseParams, err error) {
	params.address, err = cmd.Flags().GetString("address")
	if err != nil {
		return
	}
	params.srcPath, err = cmd.Flags().GetString("src")
	if err != nil {
		fmt.Println("srcPath error")
		return
	}

	fileModeRaw, err := cmd.Flags().GetUint32("mode")
	if err != nil {
		fileModeRaw = uint32(defaultFileModeRaw)
	}
	params.fileMode = os.FileMode(fileModeRaw)

	timeoutMs, err := cmd.Flags().GetUint("timeout")
	if err != nil {
		return
	}
	params.timeout = time.Duration(timeoutMs) * time.Millisecond

	return
}

func cmdOsFileReadRunE(cmd *cobra.Command, args []string) (err error) {
	params, err := getFileBaseParams(cmd)
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
		{"Destination file (host)", dstPath},
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

	hash, err := idefixgo.FileSHA256b64(ic, params.address, params.srcPath, params.timeout)
	if err != nil {
		return err
	}

	// Check hash of data
	dataHash := Sha256B64(data)
	if dataHash != hash {
		return fmt.Errorf("read error. Hash mismatch: %s != %s", dataHash, hash)
	}

	err = os.WriteFile(dstPath, data, params.fileMode)
	if err != nil {
		return err
	}

	syscall.Sync()
	return
}

func cmdOsFileWriteRunE(cmd *cobra.Command, args []string) (err error) {
	params, err := getFileBaseParams(cmd)
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
		{"Source file (host)", params.srcPath},
		{"Destination file (device)", dstPath},
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

	srcBytesHash := Sha256B64(srcBytes)

	dstBytesHash, err := idefixgo.FileWrite(ic, params.address, dstPath, srcBytes, params.fileMode, params.timeout)
	if err != nil {
		return err
	}

	if srcBytesHash != dstBytesHash {
		return fmt.Errorf("write error. Resulting hash mismatch: %s != %s", srcBytesHash, dstBytesHash)
	}

	return
}
