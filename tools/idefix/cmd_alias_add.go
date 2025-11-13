package main

import (
	"fmt"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/spf13/cobra"
)

func init() {
	cmdAliasAdd.Flags().StringP("alias", "", "", "Alias to add")
	cmdAliasAdd.MarkFlagRequired("alias")
	cmdAlias.AddCommand(cmdAliasAdd)
}

var cmdAliasAdd = &cobra.Command{
	Use:   "add",
	Short: "add client alias",
	RunE:  cmdAliasAddRunE,
}

func cmdAliasAddRunE(cmd *cobra.Command, args []string) (err error) {
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

	msg := &m.AddressAliasAddMsg{
		Address: conf.address,
		Alias:   alias,
	}

	res := &m.AddressAliasAddResponseMsg{}
	res, err = ic.AddressAliasAdd(msg)
	if err != nil {
		return fmt.Errorf("cannot add device alias: %w", err)
	}

	fmt.Printf("Alias %s added successfully\n", alias)
	fmt.Println("Current aliases:")
	for _, a := range res.Alias {
		fmt.Println(a)
	}

	return nil

}
