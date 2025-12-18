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
	cmdLog.Flags().BoolP("timestamp", "t", false, "Show timestamp (-t)")
	cmdLog.Flags().Bool("timestamp-with-delta", false, "Show timestamp with delta")
	cmdLog.Flags().BoolP("beautify", "b", false, "Enable colored output")
	rootCmd.AddCommand(cmdLog)

	cmdStream.Flags().StringP("address", "a", "", "Device address")
	cmdStream.Flags().BoolP("timestamp", "t", false, "Show timestamp (-t)")
	cmdStream.Flags().Bool("timestamp-with-delta", false, "Show timestamp with delta")

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
	timestamp, err := cmd.Flags().GetBool("timestamp")
	if err != nil {
		timestamp = false
	}
	timestampWithDelta, err := cmd.Flags().GetBool("timestamp-with-delta")
	if err != nil {
		timestampWithDelta = false
	}

	var timestampLevel int
	switch {
	case timestampWithDelta:
		timestampLevel = 2

	case timestamp:
		timestampLevel = 1

	default:
		timestampLevel = 0
	}
	useColor, err := cmd.Flags().GetBool("beautify")
	if err != nil {
		useColor = false
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

			fmt.Print(generateLogLine(timestampLevel, m, l, useColor, &lastLogTime))

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
	timestamp, err := cmd.Flags().GetBool("timestamp")
	if err != nil {
		timestamp = false
	}
	timestampWithDelta, err := cmd.Flags().GetBool("timestamp-with-delta")
	if err != nil {
		timestampWithDelta = false
	}

	var timestampLevel int
	switch {
	case timestampWithDelta:
		timestampLevel = 2

	case timestamp:
		timestampLevel = 1

	default:
		timestampLevel = 0
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
			fmt.Print(generateStreamLine(timestampLevel, k.To, k.Data, &lastLogTime))

		case <-s.Context().Done():
			return s.Context().Err()

		case <-ic.Context().Done():
			return nil
		}
	}
}

// ANSI color codes
const (
	colorRed     = 31
	colorGreen   = 32
	colorYellow  = 33
	colorMagenta = 35
	colorCyan    = 36
	colorBold    = 1
)

func colorize(s interface{}, c int, disabled bool) string {
	if disabled {
		return fmt.Sprintf("%s", s)
	}
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}

func formatLogLevel(level int) string {
	switch level {
	case -1:
		return colorize("TRC", colorMagenta, false)
	case 0:
		return colorize("DBG", colorCyan, false)
	case 1:
		return colorize("INF", colorGreen, false)
	case 2:
		return colorize("WRN", colorYellow, false)
	case 3:
		return colorize(colorize("ERR", colorRed, false), colorBold, false)
	default:
		return colorize(colorize("FTL", colorRed, false), colorBold, false)
	}
}

// formatLineWithTimestamp formats a message with optional timestamp and delta
func formatLineWithTimestamp(verboseLevel int, data string, lastLogTime *time.Time) string {
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

func generateLogLine(verboseLevel int, message string, level int, useColor bool, lastLogTime *time.Time) string {
	var prefix string
	if useColor {
		prefix = fmt.Sprintf("%s ", formatLogLevel(level))
	}

	data := fmt.Sprintf("%s%s", prefix, message)
	return formatLineWithTimestamp(verboseLevel, data, lastLogTime)
}

func generateStreamLine(verboseLevel int, to string, payload any, lastLogTime *time.Time) string {
	data := fmt.Sprintf("[%v] %v", to, payload)
	return formatLineWithTimestamp(verboseLevel, data, lastLogTime)
}
