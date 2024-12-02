package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jaracil/ei"
	"github.com/spf13/cobra"
)

// These are extracted from zerolog source to recolor the levels
const (
	// colorBlack   = 30
	colorRed    = 31
	colorGreen  = 32
	colorYellow = 33
	// colorBlue    = 34
	colorMagenta = 35
	colorCyan    = 36
	// colorWhite   = 37

	colorBold = 1
	// colorDarkGray = 90
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

func colorize(s interface{}, c int, disabled bool) string {
	if disabled {
		return fmt.Sprintf("%s", s)
	}
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
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

	s, err := ic.NewSubscriberStream(addr, "sys.evt.log", 100, false, time.Minute*10)
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
			noColor := false

			l_str := ""

			switch l {
			case -1:
				l_str = colorize("TRC", colorMagenta, noColor)
			case 0:
				l_str = colorize("DBG", colorCyan, noColor)
			case 1:
				l_str = colorize("INF", colorGreen, noColor)
			case 2:
				l_str = colorize("WRN", colorYellow, noColor)
			case 3:
				l_str = colorize(colorize("ERR", colorRed, noColor), colorBold, noColor)
			case 4:
				l_str = colorize(colorize("FTL", colorRed, noColor), colorBold, noColor)
			case 5:
				l_str = colorize(colorize("PNC", colorRed, noColor), colorBold, noColor)
			default:
				l_str = colorize("???", colorBold, noColor)
			}

			fmt.Printf("[%s] %s\n", l_str, m)

		case <-s.Context().Done():
			return s.Context().Err()

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

	s, err := ic.NewSubscriberStream(addr, args[0], 100, false, time.Minute*10)
	if err != nil {
		log.Fatalln("Cannot open stream:", err)
	}

	defer s.Close()

	fmt.Printf("-- Streaming %s %s --\n", addr, args[0])
	for {
		select {
		case k := <-s.Channel():
			fmt.Printf("%v\n", k.Data)

		case <-s.Context().Done():
			return s.Context().Err()

		case <-ic.Context().Done():
			return nil
		}
	}
}
