package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	cmdEfirm.PersistentFlags().StringP("address", "a", "", "Device address")
	cmdEfirm.MarkFlagRequired("address")

	cmdEfirmUpdate.Flags().StringP("file", "f", "", "This is the update file")
	cmdEfirmUpdate.MarkFlagRequired("file")

	cmdEfirmUpdate.Flags().StringP("dev-type", "d", "", "This is the device type. Options: rp2040")
	cmdEfirmUpdate.MarkFlagRequired("dev-type")

	cmdEfirmUpdate.Flags().StringP("usb-port", "p", "", "This is the usb tag assigned to the device's current usb path. Idefix must recognize this tag (e.g: 'P0' could be a tag for the usb path '3-1.1'). 'usb-port' and 'usb-path' are mutually exclusive and not all devices will require them")
	cmdEfirmUpdate.Flags().StringP("usb-path", "P", "", "This is the device's usb path (e.g: '3-1.1'). 'usb-port' and 'usb-path' are mutually exclusive and not all devices will require them")
	cmdEfirmUpdate.Flags().StringP("file-type", "t", "", "This is the update file type ('bin', 'uf2', 'elf', 'tar', etc.). If omitted, file's extension will be used to infer the file type. This parameter will be omitted if the file type is not recognized")
	cmdEfirmUpdate.Flags().UintP("timeout", "", 60000, "timeout in milliseconds")
	cmdEfirm.AddCommand(cmdEfirmUpdate)

	rootCmd.AddCommand(cmdEfirm)
}

var cmdEfirm = &cobra.Command{
	Use:   "efirm",
	Short: "Manage firmware of idefix's physically connected devices",
	Args:  cobra.MinimumNArgs(1),
}

var cmdEfirmUpdate = &cobra.Command{
	Use:   "update <update-file>",
	Short: "Update device with the given file",
	RunE:  cmdEfirmUpdateRunE,
}

func cmdEfirmUpdateRunE(cmd *cobra.Command, args []string) error {
	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()
	address, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
	}
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	devTypeStr, err := cmd.Flags().GetString("dev-type")
	if err != nil {
		return err
	}
	devType, err := m.ParseDevType(devTypeStr)
	if err != nil {
		return err
	}

	usbPort, err := cmd.Flags().GetString("usb-port")
	if err != nil {
		return err
	}

	usbPath, err := cmd.Flags().GetString("usb-path")
	if err != nil {
		return err
	}

	if usbPort != "" && usbPath != "" {
		return fmt.Errorf("cannot specify both 'usb-port' and 'usb-path'")
	}

	fileTypeStr, err := cmd.Flags().GetString("file-type")
	if err != nil {
		return err
	}
	var fileType m.UpdateFileType
	if fileTypeStr == "" {
		fileType = m.UPDATE_FILE_TYPE_UNSPECIFIED
	} else {
		fileType, err = m.ParseFileType(fileTypeStr)
		if err != nil {
			return err
		}
	}
	if fileType == m.UPDATE_FILE_TYPE_UNSPECIFIED {
		extIndex := strings.LastIndex(file, ".")
		if extIndex != -1 {
			fileTypeStr = file[extIndex+1:]
			fileType, err = m.ParseFileType(fileTypeStr)
			if err != nil {
				fileType = m.UPDATE_FILE_TYPE_UNSPECIFIED
			}
		}
	}
	if fileType == m.UPDATE_FILE_TYPE_UNSPECIFIED {
		fileTypeStr = "unespecified"
	}

	fileData, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	fileHash := sha256.Sum256(fileData)
	if err != nil {
		return err
	}
	dstParamsTable := pterm.TableData{
		{"Destination", ""},
		{"Address", address},
		{"Device Type", devTypeStr},
	}
	if usbPort != "" {
		dstParamsTable = append(dstParamsTable, []string{"USB Port", usbPort})
	} else if usbPath != "" {
		dstParamsTable = append(dstParamsTable, []string{"USB Path", usbPath})
	}

	pterm.DefaultTable.WithHasHeader().WithData(dstParamsTable).Render()
	fmt.Println()

	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Update", "", ""},
		{"File", file},
		{"Size", fmt.Sprintf("%d", len(fileData))},
		{"File Type", fileTypeStr},
		{"File Hash", hex.EncodeToString(fileHash[:])},
	}).Render()
	fmt.Println()

	if result, _ := pterm.DefaultInteractiveConfirm.Show(); !result {
		return nil
	}

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start("Sending update...")

	msg := m.UpdateDevFirmReqMsg{
		DevType:  devType,
		UsbPort:  usbPort,
		UsbPath:  usbPath,
		FileType: fileType,
		File:     fileData,
		FileHash: fileHash[:],
	}
	resp := m.UpdateDevFirmResMsg{}
	timeout := getTimeout(cmd)
	err = ic.Call2(address, &m.Message{To: m.TopicCmdUpdateDevFirm, Data: &msg}, &resp, timeout)
	if err != nil {
		spinner.Fail()
		return err
	}
	fmt.Println(resp.Output)
	spinner.Success()
	return nil
}
