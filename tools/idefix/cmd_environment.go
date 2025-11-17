package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nayarsystems/idefix-go/messages"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cmdEnvironment)
	cmdEnvironment.Flags().StringP("address", "a", "", "Target address. If neither the address nor the domain is indicated, the environment of the client's own address will be consulted")
	cmdEnvironment.Flags().StringP("domain", "d", "", "Target domain")
	cmdEnvironment.Flags().StringSliceP("keys", "k", []string{}, "Get specific environment keys")
}

var cmdEnvironment = &cobra.Command{
	Use:     "environment",
	Aliases: []string{"env"},
	Short:   "Get your environment info",
	RunE:    cmdEnvironmentRunE,
}

func cmdEnvironmentRunE(cmd *cobra.Command, args []string) error {
	address, _ := cmd.Flags().GetString("address")
	domain, _ := cmd.Flags().GetString("domain")
	keys, _ := cmd.Flags().GetStringSlice("keys")

	if address != "" && domain != "" {
		return fmt.Errorf("cannot specify both address and domain")
	}

	spinner, err := pterm.DefaultSpinner.WithShowTimer(true).Start("Executing request...")
	if err != nil {
		return err
	}

	env, err := getEnvironment(address, domain, keys, getTimeout(cmd))
	if err != nil {
		spinner.Fail()
		return err
	}

	rj, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		spinner.Fail()
		return err
	}
	spinner.Success()
	fmt.Printf("%s\n", rj)
	return nil
}

func getEnvironment(address string, domain string, keys []string, timeout time.Duration) (map[string]string, error) {
	envGetMsg := &messages.EnvironmentGetMsg{
		Address: address,
		Domain:  domain,
		Keys:    keys,
	}

	ic, err := getConnectedClient()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ic.Context(), timeout)
	defer cancel()
	res, err := ic.Environment(envGetMsg, ctx)
	if err != nil {
		return nil, err
	}

	if len(res.AvailableKeys) > 0 {
		res, err = ic.Environment(&messages.EnvironmentGetMsg{
			Address: address,
			Domain:  domain,
			Keys:    res.AvailableKeys,
		}, ctx)
		if err != nil {
			return nil, err
		}
	}
	return res.Environment, nil
}
