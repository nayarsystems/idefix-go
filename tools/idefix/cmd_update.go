package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gabstv/go-bsdiff/pkg/bsdiff"
	"github.com/gabstv/go-bsdiff/pkg/bspatch"
	"github.com/jaracil/ei"
	"github.com/spf13/cobra"
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

	rootCmd.AddCommand(cmdUpdate)
}

var cmdUpdate = &cobra.Command{
	Use:   "update",
	Short: "TODO update desc",
}

var cmdUpdateCreate = &cobra.Command{
	Use:   "create",
	Short: "Generate a BSDIFF4 patch",
	RunE:  cmdUpdateCreateRunE,
}

var cmdUpdateApply = &cobra.Command{
	Use:   "apply",
	Short: "TODO apply desc",
	RunE:  cmdUpdateApplyRunE,
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
