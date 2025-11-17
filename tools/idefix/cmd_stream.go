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
	cmdLog.Flags().CountP("verbose", "v", "Show timestamp (-v) or timestamp with delta (-vv)")
	rootCmd.AddCommand(cmdLog)

	cmdStream.Flags().StringP("address", "a", "", "Device address")
	cmdStream.Flags().CountP("verbose", "v", "Show timestamp (-v) or timestamp with delta (-vv)")
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
	verbose, err := cmd.Flags().GetCount("verbose")
	if err != nil {
		verbose = 0
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

	var lastLogTime time.Time

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

			fmt.Print(generateLine(verbose, m, &lastLogTime))

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
	verbose, err := cmd.Flags().GetCount("verbose")
	if err != nil {
		verbose = 0
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

	var lastLogTime time.Time

	fmt.Printf("-- Streaming %s %s --\n", addr, args[0])
	for {
		select {
		case k := <-s.Channel():
			fmt.Print(generateLine(verbose, k.Data, &lastLogTime))

		case <-s.Context().Done():
			return s.Context().Err()

		case <-ic.Context().Done():
			return nil
		}
	}
}

func generateLine(verboseLevel int, data any, lastLogTime *time.Time) string {
	var result string

	now := time.Now()
	switch verboseLevel {
	case 0:
		// No timestamp
		result = fmt.Sprintf("%v\n", data)
	case 1:
		// Only timestamp (-v)
		result = fmt.Sprintf("[%v] %v\n", now.Format("15:04:05.000"), data)
	default:
		// Timestamp + delta (-vv)
		if lastLogTime.IsZero() {
			result = fmt.Sprintf("[%v (+%v)] %v\n", now.Format("15:04:05.000"), time.Duration(0), data)
		} else {
			result = fmt.Sprintf("[%v (+%v)] %v\n", now.Format("15:04:05.000"), now.Sub(*lastLogTime), data)
		}
	}
	*lastLogTime = now

	return result
}
