package main

import (
	"fmt"
	"os"
	"syscall"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/spf13/cobra"
)

func init() {
	cmdConfigGet.Flags().StringP("file", "f", "", "destination path of received file (empty to print to stdout)")
	cmdConfig.AddCommand(cmdConfigGet)
}

var cmdConfigGet = &cobra.Command{
	Use:   "get",
	Short: "get remote client configuration",
	RunE:  cmdConfigGetRunE,
}

func cmdConfigGetRunE(cmd *cobra.Command, args []string) (err error) {
	conf := cmdConfigGetBaseParams(cmd)

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	msg := &m.AddressConfigGetMsg{
		Address: conf.address,
	}
	var res m.AddressConfigGetResponseMsg
	err = ic.Call2(m.IdefixCmdPrefix, &m.Message{To: m.CmdAddressConfigGet, Data: msg}, &res, getTimeout(cmd))
	if err != nil {
		return fmt.Errorf("cannot get client configuration: %w", err)
	}

	var file string
	file, err = cmd.Flags().GetString("file")
	if err != nil || file == "" {
		fmt.Println(string(res.Config))
		fmt.Println()
	} else {
		err = os.WriteFile(file, res.Config, 0644)
		if err != nil {
			return fmt.Errorf("cannot write to file %s: %w", file, err)
		}
		syscall.Sync()
	}
	return nil
}
