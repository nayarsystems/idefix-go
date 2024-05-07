package main

import (
	"fmt"
	"os"
	"syscall"

	idefixgo "github.com/nayarsystems/idefix-go"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {

	cmdMenhir.AddCommand(cmdMenhirStorage)

	cmdMenhirStorageRead.PersistentFlags().StringP("src", "s", "", "source file in device")
	cmdMenhirStorageRead.PersistentFlags().StringP("dst", "d", "", "destination file in host")
	cmdMenhirStorageRead.MarkPersistentFlagRequired("src")
	cmdMenhirStorageRead.MarkPersistentFlagRequired("dst")
	cmdMenhirStorage.AddCommand(cmdMenhirStorageRead)

	cmdMenhirStorageWrite.PersistentFlags().StringP("src", "s", "", "source file in host")
	cmdMenhirStorageWrite.PersistentFlags().StringP("dst", "d", "", "destination file in device")
	cmdMenhirStorageWrite.MarkPersistentFlagRequired("src")
	cmdMenhirStorageWrite.MarkPersistentFlagRequired("dst")
	cmdMenhirStorage.AddCommand(cmdMenhirStorageWrite)

	cmdMenhirStorageRemove.PersistentFlags().StringP("file", "f", "", "file to remove")
	cmdMenhirStorageRemove.MarkPersistentFlagRequired("file")
	cmdMenhirStorage.AddCommand(cmdMenhirStorageRemove)

	cmdMenhirStorageStat.PersistentFlags().StringP("file", "f", "", "file from device to get info")
	cmdMenhirStorageStat.MarkPersistentFlagRequired("file")
	cmdMenhirStorage.AddCommand(cmdMenhirStorageStat)

	cmdMenhirStorage.AddCommand(cmdMenhirStorageStats)
}

var cmdMenhirStorage = &cobra.Command{
	Use:   "storage",
	Short: "storage related commands",
}

var cmdMenhirStorageRead = &cobra.Command{
	Use:   "read",
	Short: "read file",
	RunE:  cmdMenhirStorageReadRunE,
}

var cmdMenhirStorageWrite = &cobra.Command{
	Use:   "write",
	Short: "write file",
	RunE:  cmdMenhirStorageWriteRunE,
}

var cmdMenhirStorageRemove = &cobra.Command{
	Use:   "remove",
	Short: "remove file",
	RunE:  cmdMenhirStorageRemoveRunE,
}

var cmdMenhirStorageStat = &cobra.Command{
	Use:   "stat",
	Short: "get file info",
	RunE:  cmdMenhirStorageStatRunE,
}

var cmdMenhirStorageStats = &cobra.Command{
	Use:   "stats",
	Short: "get storage stats",
	RunE:  cmdMenhirStorageStatsRunE,
}

type MenhirStorageWRParams struct {
	menhirParams
	src, dst string
}

func getMenhirStorageWRParams(cmd *cobra.Command) (params MenhirStorageWRParams, err error) {
	params.menhirParams, err = getMenhirParams(cmd)
	if err != nil {
		return
	}

	params.src, err = cmd.Flags().GetString("src")
	if err != nil {
		return
	}
	params.dst, err = cmd.Flags().GetString("dst")
	if err != nil {
		return
	}
	return
}

func cmdMenhirStorageReadRunE(cmd *cobra.Command, args []string) (err error) {
	params, err := getMenhirStorageWRParams(cmd)
	if err != nil {
		return
	}

	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Device", ""},
		{"Address", params.address},
		{"Instance", params.instance},
		{"Source file (device)", params.src},
		{"Destination file (host)", params.dst},
	}).Render()

	fmt.Println()

	if result, _ := pterm.DefaultInteractiveConfirm.Show(); !result {
		return nil
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start()

	var data []byte

	ic, err := getConnectedClient()
	if err != nil {
		goto end
	}

	data, err = idefixgo.MenhirStorageRead(ic, params.address, params.instance, params.src, params.timeout)
	if err != nil {
		goto end
	}

	err = os.WriteFile(params.dst, data, 0644)
	if err != nil {
		goto end
	}

end:
	if err != nil {
		fmt.Println(err)
		spinner.Fail()
	} else {
		syscall.Sync()
		spinner.Success()
	}
	return
}

func cmdMenhirStorageWriteRunE(cmd *cobra.Command, args []string) (err error) {
	params, err := getMenhirStorageWRParams(cmd)
	if err != nil {
		return
	}

	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Device", ""},
		{"Address", params.address},
		{"Instance", params.instance},
		{"Source file (host)", params.src},
		{"Destination file (device)", params.dst},
	}).Render()

	if result, _ := pterm.DefaultInteractiveConfirm.Show(); !result {
		return nil
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start()

	var data []byte
	ic, err := getConnectedClient()
	if err != nil {
		goto end
	}

	data, err = os.ReadFile(params.src)
	if err != nil {
		goto end
	}

	err = idefixgo.MenhirStorageWrite(ic, params.address, params.instance, params.dst, data, params.timeout)
	if err != nil {
		goto end
	}
end:
	if err != nil {
		fmt.Println(err)
		spinner.Fail()
	} else {
		syscall.Sync()
		spinner.Success()
	}
	return
}

func cmdMenhirStorageRemoveRunE(cmd *cobra.Command, args []string) (err error) {
	params, err := getMenhirParams(cmd)
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
		{"Instance", params.instance},
		{"File", file},
	}).Render()

	if result, _ := pterm.DefaultInteractiveConfirm.Show(); !result {
		return nil
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start()

	ic, err := getConnectedClient()
	if err != nil {
		goto end
	}

	err = idefixgo.MenhirStorageRemove(ic, params.address, params.instance, file, params.timeout)
	if err != nil {
		goto end
	}

end:
	if err != nil {
		fmt.Println(err)
		spinner.Fail()
	} else {
		syscall.Sync()
		spinner.Success()
	}
	return
}

func cmdMenhirStorageStatRunE(cmd *cobra.Command, args []string) (err error) {
	params, err := getMenhirParams(cmd)
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
		{"Instance", params.instance},
		{"File", file},
	}).Render()

	if result, _ := pterm.DefaultInteractiveConfirm.Show(); !result {
		return nil
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start()

	var typeStr string
	var info idefixgo.MenhirStorageFileInfo
	ic, err := getConnectedClient()
	if err != nil {
		goto end
	}

	info, err = idefixgo.MenhirStorageStat(ic, params.address, params.instance, file, params.timeout)
	if err != nil {
		goto end
	}

	switch info.Type {
	case idefixgo.MenhirStorageFileTypeRegular:
		typeStr = "Regular"
	case idefixgo.MenhirStorageFileTypeDirectory:
		typeStr = "Directory"
	default:
		typeStr = "Unknown"
	}
	fmt.Printf("File: %s\n", info.Filename)
	fmt.Printf("Type: %s\n", typeStr)
	fmt.Printf("Size: %d bytes\n", info.Size)

end:
	if err != nil {
		fmt.Println(err)
		spinner.Fail()
	} else {
		syscall.Sync()
		spinner.Success()
	}
	return
}

func cmdMenhirStorageStatsRunE(cmd *cobra.Command, args []string) (err error) {
	params, err := getMenhirParams(cmd)
	if err != nil {
		return
	}

	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Device", ""},
		{"Address", params.address},
		{"Instance", params.instance},
	}).Render()

	if result, _ := pterm.DefaultInteractiveConfirm.Show(); !result {
		return nil
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start()

	var stats idefixgo.MenhirStorageFSStats
	ic, err := getConnectedClient()
	if err != nil {
		goto end
	}

	stats, err = idefixgo.MenhirStorageStats(ic, params.address, params.instance, params.timeout)
	if err != nil {
		goto end
	}

	fmt.Printf("Block size: %d bytes\n", stats.BlockSize)
	fmt.Printf("Block count: %d\n", stats.BlockCount)
	fmt.Printf("Blocks used: %d\n", stats.BlocksUsed)

end:
	if err != nil {
		fmt.Println(err)
		spinner.Fail()
	} else {
		syscall.Sync()
		spinner.Success()
	}
	return
}
