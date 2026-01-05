package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/araddon/dateparse"
	idefix "github.com/nayarsystems/idefix-go"
	ie "github.com/nayarsystems/idefix-go/errors"
	"github.com/nayarsystems/idefix-go/messages"
)

func handleInterrupts(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt) // SIGINT

	go func() {
		<-sigChan
		cancel()
	}()
}

var client *idefix.Client

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	handleInterrupts(cancel)

	// Parse command-line flags
	var userAddress, deviceAddress, token, domain, sinceRaw, cursor string

	flag.StringVar(&userAddress, "user", "", "User address")
	flag.StringVar(&token, "token", "", "User token")
	flag.StringVar(&deviceAddress, "a", "", "Device address from which to get the events (if empty, all devices in the domain)")
	flag.StringVar(&domain, "d", "", "Domain from which to get the events (can be empty if device address is set)")
	flag.StringVar(&sinceRaw, "s", "", "(optional) Start fetching events since this date/time in formatted string (e.g., '2024-01-02 15:04:05' or '2024-01-02T15:04:05Z')")
	flag.StringVar(&cursor, "c", "", "(optional) Start fetching events from this cursor (pagination)")
	flag.Parse()

	if userAddress == "" {
		slog.Error("user address must be specified")
		os.Exit(1)
	}

	if deviceAddress == "" && domain == "" {
		slog.Error("either device address or domain must be specified")
		os.Exit(1)
	}

	var since time.Time
	if sinceRaw != "" {
		var err error
		since, err = dateparse.ParseStrict(sinceRaw)
		if err != nil && sinceRaw != "" {
			slog.Error("invalid since timestamp", "error", err)
			os.Exit(1)
		}
	}

	// Initialize Idefix client
	clientOptions := &idefix.ClientOptions{
		Broker:   "ssl://idefix.nayar.systems:8883",
		Encoding: "mg",
		Address:  userAddress,
		Token:    token,
	}

	client = idefix.NewClient(ctx, clientOptions)

	// Connect to Idefix
	err := client.Connect()
	if err != nil {
		slog.Error("failed to connect to Idefix", "error", err)
		os.Exit(1)
	}
	defer client.Disconnect()

	slog.Info("connected to Idefix")

	// Fetch and process bstates events in a loop
	for ctx.Err() == nil {
		slog.Info("fetching bstates events...", "domain", domain, "address", deviceAddress, "since", since, "cursor", cursor)
		res, err := client.EventsGet(&messages.EventsGetMsg{
			Type:           "application/vnd.nayar.bstates", // bstates events only
			Address:        deviceAddress,                   // events from this device address only (if empty, all addresses in the specified domain)
			Domain:         domain,                          // events from this domain only
			Since:          since,                           // fetch events since this time (can be empty and it's ignored if cursor is set)
			ContinuationID: cursor,                          // pagination cursor (can be empty)
			Timeout:        time.Minute,                     // 1 minute of long polling timeout. If no events arrive in this time, the request will timeout and we will re-issue it
		})

		if err != nil {
			if !ie.ErrTimeout.Is(err) {
				slog.Error("failed to get events", "error", err)
				os.Exit(1)
			}
			// Request timed out, re-issue it
			continue
		}

		// Update cursor for next fetch
		cursor = res.ContinuationID

		// Process events
		for _, event := range res.Events {
			if err := processBstatesEvent(event); err != nil {
				slog.Error("failed to process bstates event", "id", event.UID, "error", err)
			}
			if ctx.Err() != nil {
				break
			}
		}
	}
	slog.Info("shutting down")
}
