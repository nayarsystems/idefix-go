package main

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/jaracil/ei"
	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	cmdUsb.PersistentFlags().StringP("address", "a", "", "Device address")

	cmdUsbGet.Flags().StringP("field-filter", "f", "^(PATH|A:|E:)", "Device will only send usb info fields that match with the given regular expression")
	cmdUsb.AddCommand(cmdUsbGet)

	rootCmd.AddCommand(cmdUsb)
}

var cmdUsb = &cobra.Command{
	Use:   "usb",
	Short: "Manage idefix's connected usb devices",
	Args:  cobra.MinimumNArgs(1),
}

var cmdUsbGet = &cobra.Command{
	Use:   "get",
	Short: "Get list of connected usb devices",
	RunE:  cmdUsbGetRunE,
}

func cmdUsbGetRunE(cmd *cobra.Command, args []string) error {
	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()
	address, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
	}
	fieldFilter, err := cmd.Flags().GetString("field-filter")
	if err != nil {
		return err
	}

	actualFieldFilter := fmt.Sprintf("(%s|^USB_)", fieldFilter)
	_, err = regexp.Compile(actualFieldFilter)
	if err != nil {
		return fmt.Errorf("error parsing field filter: %v", err)
	}

	msg := m.DevListReqMsg{
		Expr:        "{\"$exists\":\"USB_PATH\"}",
		FieldFilter: actualFieldFilter,
	}

	timeout := getTimeout(cmd)

	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start(fmt.Sprintf("Querying usb devices connected to %s...", address))

	resp := &m.DevListResponseMsg{}
	err = ic.Call2(address, &m.Message{To: m.TopicCmdGetDevInfo, Data: &msg}, resp, timeout)
	if err != nil {
		spinner.Fail()
		return err
	}
	devices := map[string]map[string]any{}
	for _, devEnv := range resp.Devices {
		usbPath := ei.N(devEnv).M("USB_PATH").StringZ()
		usbPort := ei.N(devEnv).M("USB_PORT").StringZ()
		if usbPath == "" {
			spinner.Fail()
			return fmt.Errorf("fix me: no usb path in device env")
		}
		var devId string
		if usbPort != "" {
			devId = fmt.Sprintf("%s (%s)", usbPath, usbPort)
		} else {
			devId = usbPath
		}
		delete(devEnv, "USB_PATH")
		delete(devEnv, "USB_PORT")
		devices[devId] = devEnv

	}
	rj, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		spinner.Fail()
		return err
	}
	fmt.Printf("%s\n", rj)
	spinner.Success()
	return nil
}
