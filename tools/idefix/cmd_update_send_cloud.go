package main

import (
	"fmt"
	"os"
	"time"

	idefixgo "github.com/nayarsystems/idefix-go"
	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/spf13/cobra"
)

type gsr2mgrResponse struct {
	Hash    string `json:"hash"`
	Version string `json:"version"`
	Arch    string `json:"arch"`
}

func cmdUpdateSendCloudRunE(cmd *cobra.Command, args []string) error {
	p, err := getUpdateParams(cmd)
	if err != nil {
		return err
	}

	if p.address == "" {
		return fmt.Errorf("address is required")
	}

	updatefile, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	updatebytes, err := os.ReadFile(updatefile)
	if err != nil {
		return err
	}

	dsthash := Sha256Hex(updatebytes)

	ic, err := getConnectedClient()
	if err != nil {
		return fmt.Errorf("cannot connect to the server: %w", err)
	}
	msg := map[string]any{
		"hash": dsthash,
	}

	var ret gsr2mgrResponse
	err = ic.Call2("gsr2mgr", &m.Message{To: "release.exists", Data: msg}, &ret, p.tout)
	if err != nil {
		return fmt.Errorf("gsr2mgr failure checking if exists: %w", err)
	}

	if ret.Version == "" {
		err := gsr2mgrUploadBinary(ic, updatebytes, dsthash)
		if err != nil {
			return fmt.Errorf("gsr2mgr failure uploading: %w", err)
		}
	}

	err = ic.Call2("gsr2mgr", &m.Message{To: "release.exists", Data: msg}, &ret, p.tout)
	if err != nil {
		return fmt.Errorf("gsr2mgr failure checking if exists: %w", err)
	}

	if ret.Version == "" {
		return fmt.Errorf("gsr2mgr failure: no version found after upload")
	}

	ic.AddressEnvironmentSet(&m.AddressEnvironmentSetMsg{
		Address:     p.address,
		Environment: map[string]string{"gsr2mgr_idefix_version": ret.Version},
	})

	resp, err := ic.Call("gsr2mgr", &m.Message{To: "device.alive", Data: map[string]any{"address": p.address}}, p.tout)
	if err != nil {
		return fmt.Errorf("gsr2mgr failure: %w", err)
	}
	if resp.Err != "" {
		return fmt.Errorf("gsr2mgr failure: %s", resp.Err)
	}

	return nil
}

func gsr2mgrUploadBinary(ic *idefixgo.Client, updatebytes []byte, dsthash string) error {
	fmt.Println("Uploading binary to gsr2mgr")
	msg := map[string]any{
		"data": updatebytes,
		"hash": dsthash,
	}

	resp, err := ic.Call("gsr2mgr", &m.Message{To: "release.upload", Data: msg}, time.Minute*2)
	if err != nil {
		return fmt.Errorf("gsr2mgr failure uploading: %w", err)
	}
	if resp.Err != "" {
		return fmt.Errorf("gsr2mgr failure uploading: %s", resp.Err)
	}

	return nil
}
