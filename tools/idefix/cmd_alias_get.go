package main

import (
	"fmt"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/spf13/cobra"
)

func init() {
	cmdAlias.AddCommand(cmdAliasGet)
}

var cmdAliasGet = &cobra.Command{
	Use:   "get",
	Short: "get client alias",
	RunE:  cmdAliasGetRunE,
}

func cmdAliasGetRunE(cmd *cobra.Command, args []string) (err error) {
	conf := cmdAliasParams(cmd)

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	msg := &m.AddressAliasGetMsg{
		Address: conf.address,
	}

	var res *m.AddressAliasGetResponseMsg
	res, err = ic.AddressAliasGet(msg)
	if err != nil {
		return fmt.Errorf("cannot get device alias: %w", err)
	}

	for _, alias := range res.Alias {
		fmt.Println(alias)
	}

	return nil
}
