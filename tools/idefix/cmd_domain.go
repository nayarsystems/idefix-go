package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"atomicgo.dev/cursor"
	ie "github.com/nayarsystems/idefix-go/errors"
	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	cmdDomainCreate.Flags().StringP("domain", "n", "", "Domain Name")
	cmdDomainCreate.Flags().String("allow", "", "Allow rule")
	cmdDomainCreate.Flags().String("deny", "", "Deny rule")
	cmdDomainCreate.Flags().String("env", "", "Environment map")
	cmdDomainCreate.Flags().StringArray("admin", []string{}, "Admin address (can be set multiple times for multiple admins)")
	cmdDomainCreate.MarkFlagRequired("domain")
	cmdDomain.AddCommand(cmdDomainCreate)

	cmdDomainAssign.Flags().StringP("domain", "n", "", "Domain Name")
	cmdDomainAssign.Flags().StringP("address", "a", "", "Device address")
	cmdDomainAssign.MarkFlagRequired("domain")
	cmdDomainAssign.MarkFlagRequired("address")
	cmdDomain.AddCommand(cmdDomainAssign)

	cmdDomainGet.Flags().StringP("domain", "n", "", "Domain Name")
	cmdDomainGet.Flags().StringP("address", "a", "", "Address to obtain domain from")
	cmdDomainGet.MarkFlagsMutuallyExclusive("domain", "address")
	cmdDomain.AddCommand(cmdDomainGet)

	cmdDomainDelete.Flags().StringP("domain", "n", "", "Domain Name")
	cmdDomainDelete.MarkFlagRequired("domain")
	cmdDomain.AddCommand(cmdDomainDelete)

	cmdDomainUpdate.Flags().StringP("domain", "d", "", "Domain Name")
	cmdDomainUpdate.Flags().String("allow", "", "Allow rule")
	cmdDomainUpdate.Flags().String("deny", "", "Deny rule")
	cmdDomainUpdate.Flags().String("env", "", "Environment Map")
	cmdDomainUpdate.Flags().StringArray("admin", []string{}, "Admin Address (Can be set multiple times)")
	cmdDomainUpdate.MarkFlagRequired("domain")
	cmdDomain.AddCommand(cmdDomainUpdate)

	// Shared flags (persistent so both "get" and "watch" inherit them).
	cmdDomainAddresses.PersistentFlags().StringP("domain", "n", "", "Domain Name")
	cmdDomainAddresses.PersistentFlags().Uint("limit", 0, "Max addresses per page")
	cmdDomainAddresses.PersistentFlags().String("cid", "", "Continuation id to resume a sync")
	cmdDomainAddresses.PersistentFlags().StringSlice("fields", []string{}, "Client fields to return (dot notation: domain, aliases, env.<k>, lastState.<src>...)")
	cmdDomainAddresses.MarkPersistentFlagRequired("domain")

	cmdDomainAddressesWatch.Flags().String("timeout", "20s", "Long-poll wait for changes")
	cmdDomainAddressesWatch.Flags().Bool("changes-only", false, "Skip the initial snapshot output; only print changes")

	cmdDomain.AddCommand(cmdDomainAddresses)
	cmdDomainAddresses.AddCommand(cmdDomainAddressesGet)
	cmdDomainAddresses.AddCommand(cmdDomainAddressesWatch)

	rootCmd.AddCommand(cmdDomain)
}

var cmdDomain = &cobra.Command{
	Use:     "domain",
	Aliases: []string{"domains"},
	Short:   "Manage idefix domains",
}

var cmdDomainGet = &cobra.Command{
	Use:   "get",
	Short: "Get a domain",
	RunE:  cmdDomainGetRunE,
}

var cmdDomainDelete = &cobra.Command{
	Use:   "delete",
	Short: "Delete a domain",
	RunE:  cmdDomainDeleteRunE,
}

var cmdDomainUpdate = &cobra.Command{
	Use:   "update",
	Short: "Update a domain. Any field specified here will totally overwrite the current value (won't be appended)",
	RunE:  cmdDomainUpdateRunE,
}

var cmdDomainAssign = &cobra.Command{
	Use:   "assign",
	Short: "Assing a device to a domain",
	RunE:  cmdDomainAssignRunE,
}

var cmdDomainCreate = &cobra.Command{
	Use:   "create",
	Short: "Create a domain. If you dont specify a domain administrator, your address will be set as administrator of the new domain",
	RunE:  cmdDomainCreateRunE,
}

var cmdDomainAddresses = &cobra.Command{
	Use:     "addresses",
	Aliases: []string{"address", "addrs", "addr", "a"},
	Short:   "Get or watch client fields per address in a domain",
}

var cmdDomainAddressesGet = &cobra.Command{
	Use:   "get",
	Short: "Fetch the current snapshot of a domain's addresses",
	RunE:  cmdDomainAddressesGetRunE,
}

var cmdDomainAddressesWatch = &cobra.Command{
	Use:   "watch",
	Short: "Stream a domain's addresses and their changes until cancelation",
	RunE:  cmdDomainAddressesWatchRunE,
}

// printAddressesBatch prints one batch of the stream: a header (short timestamp and the number of
// devices) on stderr, and the JSON on stdout, so stdout stays a clean JSON stream (pipeable) while
// the header still shows on a terminal right above its batch.
func printAddressesBatch(addresses map[string]any) {
	noun := "devices"
	if len(addresses) == 1 {
		noun = "device"
	}
	header := pterm.NewStyle(pterm.FgCyan, pterm.Bold).Sprintf("── %s · %d %s ──",
		time.Now().Format(time.Stamp), len(addresses), noun)
	fmt.Fprintln(os.Stderr, header)
	rj, _ := json.MarshalIndent(addresses, "", "  ")
	fmt.Printf("%s\n", rj)
}

func cmdDomainAddressesGetRunE(cmd *cobra.Command, args []string) error {
	return runDomainAddresses(cmd, false, false)
}

func cmdDomainAddressesWatchRunE(cmd *cobra.Command, args []string) error {
	changesOnly, _ := cmd.Flags().GetBool("changes-only")
	return runDomainAddresses(cmd, true, changesOnly)
}

// runDomainAddresses drives domain.get.addresses: it paginates the initial snapshot and, when
// streaming, keeps long-polling for deltas until cancelation. See the response semantics in
// messages.DomainGetAddressesResponseMsg (cid never empty, upToDate once on snapshot completion).
func runDomainAddresses(cmd *cobra.Command, stream, changesOnly bool) error {
	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	msg := &m.DomainGetAddressesMsg{}
	msg.Domain, _ = cmd.Flags().GetString("domain")
	msg.Limit, _ = cmd.Flags().GetUint("limit")
	msg.Cid, _ = cmd.Flags().GetString("cid")
	msg.Fields, _ = cmd.Flags().GetStringSlice("fields")
	if stream {
		traw, _ := cmd.Flags().GetString("timeout")
		msg.Timeout, err = time.ParseDuration(traw)
		if err != nil {
			return fmt.Errorf("invalid timeout: %w", err)
		}
	}

	// pterm's SIGINT handler shows the cursor via the cursor package's target (os.Stdout by
	// default); point it at stderr so a Ctrl-C never writes escapes into the JSON stream.
	cursor.SetTarget(os.Stderr)

	sawUpToDate := false
	for {
		// Spin while waiting, like the other commands. It goes to stderr and is removed on arrival,
		// so stdout stays a clean JSON stream. The text reflects the phase.
		text := fmt.Sprintf("Fetching addresses of %s…", msg.Domain)
		if sawUpToDate {
			text = fmt.Sprintf("Watching %s for changes…", msg.Domain)
		}
		spinner, _ := pterm.DefaultSpinner.WithWriter(os.Stderr).WithRemoveWhenDone(true).WithShowTimer(true).Start(text)

		ctx, cancelc := context.WithTimeout(ic.Context(), getTimeout(cmd)+msg.Timeout)
		res, err := ic.DomainGetAddresses(msg, ctx)
		cancelc()
		spinner.Stop()
		if err != nil {
			if rootctx.Err() == context.Canceled {
				return nil // Ctrl-C mid-call
			}
			if stream && ie.ErrTimeout.Is(err) {
				continue // transient: keep streaming
			}
			return err
		}

		// With --changes-only, suppress everything up to and including the snapshot's last page.
		if len(res.Addresses) > 0 && (!changesOnly || sawUpToDate) {
			printAddressesBatch(res.Addresses)
		}

		msg.Cid = res.Cid // the response cid is never empty; keep resending it
		if res.UpToDate {
			sawUpToDate = true
			if !stream {
				return nil // snapshot complete (get)
			}
		}

		if rootctx.Err() != nil { // Ctrl-C between calls
			if rootctx.Err() == context.Canceled {
				return nil
			}
			return rootctx.Err()
		}
	}
}

func cmdDomainCreateRunE(cmd *cobra.Command, args []string) error {
	domain, err := parseDomainFlags(cmd)
	if err != nil {
		return err
	}
	msg := &m.DomainCreateMsg{
		Domain:      domain.Domain,
		Env:         domain.Env,
		AccessRules: domain.AccessRules,
	}
	return commandCall2(m.IdefixCmdPrefix, m.CmdDomainCreate, msg, getTimeout(cmd))
}

func cmdDomainAssignRunE(cmd *cobra.Command, args []string) (err error) {
	msg := &m.DomainAssignMsg{}
	msg.Domain, err = cmd.Flags().GetString("domain")
	if err != nil {
		return err
	}

	msg.Address, err = cmd.Flags().GetString("address")
	if err != nil {
		return err
	}

	return commandCall2(m.IdefixCmdPrefix, m.CmdDomainAssign, msg, getTimeout(cmd))
}

func cmdDomainUpdateRunE(cmd *cobra.Command, args []string) error {
	domain, err := parseDomainFlags(cmd)
	if err != nil {
		return err
	}
	msg := &m.DomainUpdateMsg{
		Domain:      domain.Domain,
		Env:         domain.Env,
		AccessRules: domain.AccessRules,
	}
	return commandCall2(m.IdefixCmdPrefix, m.CmdDomainUpdate, msg, getTimeout(cmd))
}

func parseDomainFlags(cmd *cobra.Command) (domain *m.Domain, err error) {
	domain = &m.Domain{}
	domain.Domain, err = cmd.Flags().GetString("domain")
	if err != nil {
		return nil, err
	}

	env := make(map[string]string)
	senv, err := cmd.Flags().GetString("env")
	if err != nil {
		return nil, err
	}
	if cmd.Flags().Changed("env") {
		err = json.Unmarshal([]byte(senv), &env)
		if err != nil {
			return nil, fmt.Errorf("cannot parse environment: %w", err)
		}
		domain.Env = env
	}

	// TODO: Parse access rules

	return domain, nil
}

func cmdDomainGetRunE(cmd *cobra.Command, args []string) (err error) {
	domain, err := cmd.Flags().GetString("domain")
	if err == nil && domain != "" {
		msg := &m.DomainGetMsg{Domain: domain}
		return commandCall2(m.IdefixCmdPrefix, m.CmdDomainGet, msg, getTimeout(cmd))
	}
	address, err := cmd.Flags().GetString("address")
	if err == nil && address != "" {
		msg := &m.AddressDomainGetMsg{Address: address}
		return commandCall2(m.IdefixCmdPrefix, m.CmdAddressDomainGet, msg, getTimeout(cmd))
	}
	return err
}

func cmdDomainDeleteRunE(cmd *cobra.Command, args []string) (err error) {
	msg := &m.DomainDeleteMsg{}
	name, err := cmd.Flags().GetString("domain")
	if err != nil {
		return err
	}
	fmt.Println("You are about to delete the domain:", name)
	if result, _ := pterm.DefaultInteractiveConfirm.Show(); !result {
		return nil
	}
	if err := commandCall2(m.IdefixCmdPrefix, m.CmdDomainGet, msg, getTimeout(cmd)); err != nil {
		return err
	}

	return commandCall2(m.IdefixCmdPrefix, m.CmdDomainDelete, msg, getTimeout(cmd))
}
