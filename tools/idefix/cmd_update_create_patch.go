package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gabstv/go-bsdiff/pkg/bsdiff"
	"github.com/nayarsystems/idefix-go/normalize"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

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

		os.WriteFile(out, j, 0644)
	}
	pterm.Success.Println("Patch created!")

	return nil
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
