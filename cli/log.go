package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jaracil/ei"
)

type LogCmd struct {
	DeviceAddress string `name:"device" short:"d" help:"Device address" required:""`
	Wait          bool   `name:"wait" short:"w" help:"Wait if device is not connected"`
	Level         int    `name:"loglevel" short:"l" default:"1" help:"Filter lower log levels"`
}

func (logcmd *LogCmd) Run(kctx *Context) error {
	ic := kctx.Client

	err := ic.Connect()
	if err != nil {
		log.Fatalln("Cannot login:", err)
	}
	defer ic.Disconnect()

	s, err := ic.NewStream(logcmd.DeviceAddress, "sys.evt.log", 100, time.Minute*10)
	if err != nil {
		log.Fatalln("Cannot open stream:", err)
	}

	fmt.Printf("-- Streaming %s sys.evt.log --\n", logcmd.DeviceAddress)
	for {
		select {
		case k := <-s.Channel():
			l, err := ei.N(k.Data).M("level").Int()
			if err != nil {
				fmt.Println(err)
				continue
			}

			if logcmd.Level > l {
				continue
			}

			m, err := ei.N(k.Data).M("message").String()
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Printf("[%d] %s\n", l, m)

		case <-kctx.Client.Context().Done():
			return nil
		}
	}
}
