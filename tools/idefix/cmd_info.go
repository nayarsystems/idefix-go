package main

import (
	"fmt"
	"log"
	"time"

	idf "gitlab.com/garagemakers/idefix-go"
)

type InfoCmd struct {
	DeviceAddress string `name:"device" short:"d" help:"Device address" required:"" `
}

func (infocmd *InfoCmd) Run(kctx *Context) error {
	ic := kctx.Client

	err := ic.Connect()
	if err != nil {
		log.Fatalln("Cannot login:", err)
	}
	defer ic.Disconnect()

	info, err := ic.Call(infocmd.DeviceAddress, &idf.Message{To: "sys.cmd.info"}, time.Second)
	if err != nil {
		return fmt.Errorf("Cannot get the device info: %w", err)
	}

	data, ok := info.Data.(map[string]any)
	if !ok {
		return fmt.Errorf("Unexpected response: %v", info.Data)
	}

	for k, v := range data {
		fmt.Printf("%s: %v\n", k, v)
	}

	return nil
}
