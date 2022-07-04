package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jaracil/ei"
	"github.com/spf13/cobra"
)

func init() {
	cmdLog.Flags().StringP("device", "d", "", "Device ID")
	cmdLog.MarkFlagRequired("device")
	cmdLog.Flags().BoolP("wait", "w", false, "Wait for the device if not connected")
	cmdLog.Flags().IntP("loglevel", "l", 2, "Filter lower log levels")

	rootCmd.AddCommand(cmdLog)
}

var cmdLog = &cobra.Command{
	Use:   "log",
	Short: "Stream Device ID sys.evt.log messages",
	RunE:  cmdLogRunE,
}

func cmdLogRunE(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("device")
	if err != nil {
		return err
	}
	level, err := cmd.Flags().GetInt("loglevel")
	if err != nil {
		return err
	}

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	s, err := ic.NewStream(addr, "sys.evt.log", 100, time.Minute*10)
	if err != nil {
		log.Fatalln("Cannot open stream:", err)
	}

	fmt.Printf("-- Streaming %s sys.evt.log --\n", addr)
	for {
		select {
		case k := <-s.Channel():
			l, err := ei.N(k.Data).M("level").Int()
			if err != nil {
				fmt.Println(err)
				continue
			}

			if level > l {
				continue
			}

			m, err := ei.N(k.Data).M("message").String()
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Printf("[%d] %s\n", l, m)

		case <-ic.Context().Done():
			return nil
		}
	}
}
