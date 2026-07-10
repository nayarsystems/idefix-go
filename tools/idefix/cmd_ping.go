package main

import (
	"fmt"
	"math"
	"time"

	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/spf13/cobra"
)

func init() {
	cmdPing.Flags().StringP("address", "a", "", "Device address")
	cmdPing.Flags().UintP("count", "n", 0, "Stop after sending this many pings (0 = no limit, stop with Ctrl+C)")
	cmdPing.Flags().DurationP("interval", "i", time.Second, "Wait interval between sending each ping")
	cmdPing.Flags().DurationP("deadline", "w", 0, "Stop after this much time has elapsed (0 = no limit)")
	cmdPing.Flags().BoolP("quiet", "q", false, "Only print the summary at the end, not each ping result")
	cmdPing.MarkFlagRequired("address")

	rootCmd.AddCommand(cmdPing)
}

var cmdPing = &cobra.Command{
	Use:   "ping",
	Short: "Ping a device repeatedly and measure round-trip time",
	RunE:  cmdPingRunE,
}

type pingStats struct {
	transmitted uint
	received    uint
	rtts        []float64 // milliseconds
}

func (s *pingStats) record(rtt time.Duration) {
	s.received++
	s.rtts = append(s.rtts, float64(rtt.Microseconds())/1000.0)
}

func (s *pingStats) print(address string) {
	loss := float64(0)
	if s.transmitted > 0 {
		loss = 100 * float64(s.transmitted-s.received) / float64(s.transmitted)
	}

	fmt.Printf("\n--- %s ping statistics ---\n", address)
	fmt.Printf("%d packets transmitted, %d received, %.1f%% packet loss\n", s.transmitted, s.received, loss)

	if len(s.rtts) == 0 {
		return
	}

	min, max, sum := s.rtts[0], s.rtts[0], float64(0)
	for _, v := range s.rtts {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}
	avg := sum / float64(len(s.rtts))

	var sqDiffSum float64
	for _, v := range s.rtts {
		d := v - avg
		sqDiffSum += d * d
	}
	mdev := math.Sqrt(sqDiffSum / float64(len(s.rtts)))

	fmt.Printf("rtt min/avg/max/mdev = %.3f/%.3f/%.3f/%.3f ms\n", min, avg, max, mdev)
}

func cmdPingRunE(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
	}

	count, err := cmd.Flags().GetUint("count")
	if err != nil {
		return err
	}

	interval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		return err
	}

	deadline, err := cmd.Flags().GetDuration("deadline")
	if err != nil {
		return err
	}

	quiet, err := cmd.Flags().GetBool("quiet")
	if err != nil {
		return err
	}

	pingTimeout := getTimeout(cmd)

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	var deadlineCh <-chan time.Time
	if deadline > 0 {
		deadlineCh = time.After(deadline)
	}

	fmt.Printf("PING %s (%s)\n", addr, m.TopicCmdPing)

	stats := &pingStats{}

	for seq := uint32(1); count == 0 || uint(seq) <= count; seq++ {
		select {
		case <-rootctx.Done():
			stats.print(addr)
			return nil
		case <-deadlineCh:
			stats.print(addr)
			return nil
		default:
		}

		stats.transmitted++

		req := &m.PingReqMsg{Seq: seq}
		res := &m.PingResMsg{}

		t0 := time.Now()
		callErr := ic.Call2(addr, &m.Message{To: m.TopicCmdPing, Data: req}, res, pingTimeout)
		rtt := time.Since(t0)

		if !quiet {
			if callErr != nil {
				fmt.Printf("no response from %s: seq=%d (%s)\n", addr, seq, callErr)
			} else {
				fmt.Printf("pong from %s: seq=%d time=%.3fms\n", addr, res.Seq, float64(rtt.Microseconds())/1000.0)
			}
		}

		if callErr == nil {
			stats.record(rtt)
		}

		if count != 0 && uint(seq) >= count {
			break
		}

		sleep := interval - rtt
		if sleep < 0 {
			sleep = 0
		}

		select {
		case <-rootctx.Done():
			stats.print(addr)
			return nil
		case <-deadlineCh:
			stats.print(addr)
			return nil
		case <-time.After(sleep):
		}
	}

	stats.print(addr)
	return nil
}
