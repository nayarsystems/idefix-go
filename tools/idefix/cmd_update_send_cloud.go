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

	gsr2mgr, err := cmd.Flags().GetString("gsr2mgr")
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

	forgetLastUpdate, err := cmd.Flags().GetBool("forget-last-update")
	if err != nil {
		return err
	}

	hashAsVersion, err := cmd.Flags().GetBool("hash-as-version")
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
	err = ic.Call2(gsr2mgr, &m.Message{To: "release.exists", Data: msg}, &ret, p.tout)
	if err != nil {
		return fmt.Errorf("gsr2mgr failure checking if exists: %w", err)
	}

	if ret.Version == "" {
		fmt.Printf("Uploading binary with hash %s\n", dsthash)
		err := gsr2mgrUploadBinary(ic, gsr2mgr, updatebytes, dsthash)
		if err != nil {
			return fmt.Errorf("gsr2mgr failure uploading: %w", err)
		}
		fmt.Printf("Binary uploaded to gsr2mgr with hash %s\n", dsthash)
	} else {
		fmt.Printf("Binary with hash %s already exists on gsr2mgr\n", dsthash)
	}

	err = ic.Call2(gsr2mgr, &m.Message{To: "release.exists", Data: msg}, &ret, p.tout)
	if err != nil {
		return fmt.Errorf("gsr2mgr failure checking release existence: %w", err)
	}

	if ret.Version == "" {
		return fmt.Errorf("gsr2mgr failure: no version found after upload")
	}
	version := ret.Version
	if hashAsVersion {
		version = dsthash
	}

	fmt.Printf("Setting %s target version for %s to %s\n", p.targetStr, p.address, version)

	switch p.target {
	case m.LauncherTargetExec:
		if forgetLastUpdate {
			ic.AddressEnvironmentUnset(&m.AddressEnvironmentUnsetMsg{
				Address: p.address,
				Keys:    []string{"gsr2mgr_launcher_last_update"},
			})
		}
		ic.AddressEnvironmentSet(&m.AddressEnvironmentSetMsg{
			Address:     p.address,
			Environment: map[string]string{"gsr2mgr_launcher_version": version},
		})
	case m.IdefixTargetExec:
		if forgetLastUpdate {
			ic.AddressEnvironmentUnset(&m.AddressEnvironmentUnsetMsg{
				Address: p.address,
				Keys:    []string{"gsr2mgr_idefix_last_update"},
			})
		}
		ic.AddressEnvironmentSet(&m.AddressEnvironmentSetMsg{
			Address:     p.address,
			Environment: map[string]string{"gsr2mgr_idefix_version": version},
		})
	default:
		return fmt.Errorf("unknown target %d", p.target)
	}

	resp, err := ic.Call(gsr2mgr, &m.Message{To: "device.alive", Data: map[string]any{"address": p.address}}, p.tout)
	if err != nil {
		return fmt.Errorf("gsr2mgr failure: %w", err)
	}
	if resp.Err != "" {
		return fmt.Errorf("gsr2mgr failure: %s", resp.Err)
	}

	fmt.Printf("%s target version of %s successfully set to %s. Device will be updated (if it isn't already) as soon as possible\n", p.targetStr, p.address, version)
	return nil
}

func gsr2mgrUploadBinary(ic *idefixgo.Client, gsr2mgr string, updatebytes []byte, dsthash string) error {
	msg := map[string]any{
		"data": updatebytes,
		"hash": dsthash,
	}

	resp, err := ic.Call(gsr2mgr, &m.Message{To: "release.upload", Data: msg}, time.Minute*2)
	if err != nil {
		return fmt.Errorf("gsr2mgr failure uploading: %w", err)
	}
	if resp.Err != "" {
		return fmt.Errorf("gsr2mgr failure uploading: %s", resp.Err)
	}

	return nil
}
