package main

import (
	"fmt"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/spf13/cobra"
)

func init() {
	cmdAliasRemove.Flags().StringP("alias", "", "", "Alias to remove")
	cmdAliasRemove.MarkFlagRequired("alias")
	cmdAlias.AddCommand(cmdAliasRemove)
}

var cmdAliasRemove = &cobra.Command{
	Use:   "remove",
	Short: "remove client alias",
	RunE:  cmdAliasRemoveRunE,
}

func cmdAliasRemoveRunE(cmd *cobra.Command, args []string) (err error) {
	conf := cmdAliasParams(cmd)

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	alias, err := cmd.Flags().GetString("alias")
	if err != nil {
		return err
	}

	msg := &m.AddressAliasRemoveMsg{
		Address: conf.address,
		Alias:   alias,
	}

	res := &m.AddressAliasRemoveResponseMsg{}
	res, err = ic.AddressAliasRemove(msg)
	if err != nil {
		return fmt.Errorf("cannot remove device alias: %w", err)
	}

	fmt.Printf("Alias %s removed successfully\n", alias)
	fmt.Println("Current aliases:")
	for _, a := range res.Alias {
		fmt.Println(a)
	}

	return nil

}
