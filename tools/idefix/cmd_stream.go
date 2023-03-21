package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jaracil/ei"
	"github.com/spf13/cobra"
)

func init() {
	cmdLog.Flags().StringP("address", "a", "", "Device address")
	cmdLog.MarkFlagRequired("address")
	cmdLog.Flags().BoolP("wait", "w", false, "Wait for the device if not connected")
	cmdLog.Flags().IntP("loglevel", "l", 2, "Filter lower log levels")
	rootCmd.AddCommand(cmdLog)

	cmdStream.Flags().StringP("address", "a", "", "Device address")
	rootCmd.AddCommand(cmdStream)
}

var cmdLog = &cobra.Command{
	Use:   "log",
	Short: "Stream Device ID sys.evt.log messages",
	RunE:  cmdLogRunE,
}

var cmdStream = &cobra.Command{
	Use:   "stream",
	Short: "Stream device messages",
	RunE:  cmdStreamRunE,
}

func cmdLogRunE(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
	}
	level, err := cmd.Flags().GetInt("loglevel")
	if err != nil {
		level = 2
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

	defer s.Close()

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

func cmdStreamRunE(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
	}

	if len(args) != 1 {
		return fmt.Errorf("stream only supports one topic")
	}

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	s, err := ic.NewStream(addr, args[0], 100, time.Minute*10)
	if err != nil {
		log.Fatalln("Cannot open stream:", err)
	}

	defer s.Close()

	fmt.Printf("-- Streaming %s %s --\n", addr, args[0])
	for {
		select {
		case k := <-s.Channel():
			fmt.Printf("%v\n", k.Data)

		case <-ic.Context().Done():
			return nil
		}
	}
}
