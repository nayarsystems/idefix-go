package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/araddon/dateparse"
	idefix "github.com/nayarsystems/idefix-go"
	"github.com/nayarsystems/idefix-go/eventpipe"
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

	// Set slog default logger
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Parse command-line flags
	var userAddress, deviceAddress, token, domain, eventType, sinceRaw, pipeId, cursor string
	var computeStats bool

	flag.StringVar(&userAddress, "user", "", "User address")
	flag.StringVar(&token, "token", "", "User token")
	flag.StringVar(&deviceAddress, "a", "", "Device address from which to get the events (if empty, all devices in the domain)")
	flag.StringVar(&domain, "d", "", "Domain from which to get the events (can be empty if device address is set)")
	flag.StringVar(&eventType, "t", "", "Event type to fetch (default: not set, all types)")
	flag.StringVar(&sinceRaw, "s", "", "(optional) Start fetching events since this date/time in formatted string (e.g., '2024-01-02 15:04:05' or '2024-01-02T15:04:05Z')")
	flag.StringVar(&cursor, "c", "", "(optional) Start fetching events from this cursor (pagination)")
	flag.StringVar(&pipeId, "pipe", "example-events-pipe", "ID of the pipe")
	flag.BoolVar(&computeStats, "stats", false, "Show statistics")
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

	eventSourceManager, err := eventpipe.NewEventSourceManager(eventpipe.EventSourceManagerParams{
		Client:      client,
		Context:     ctx,
		Logger:      slog.Default(),
		StoragePath: "example.db",
	})
	if err != nil {
		slog.Error("failed to create event source manager", "error", err)
		os.Exit(1)
	}

	if err = eventSourceManager.Init(); err != nil {
		slog.Error("failed to initialize event source manager", "error", err)
		os.Exit(1)
	}
	defer eventSourceManager.Close()

	eventSource, err := eventSourceManager.NewSource(eventpipe.EventSourceParams{
		Id:             "example-events-pipe",
		Domain:         domain,
		Address:        deviceAddress,
		Type:           eventType,
		Since:          since,
		ContinuationID: cursor,
	})
	if err != nil {
		slog.Error("failed to create event source", "error", err)
		os.Exit(1)
	}

	stageDuration := 2 * time.Second
	if err = eventSource.Push(&mockStage{index: 0, duration: stageDuration}, eventpipe.OptName("mock 0"), eventpipe.OptConcurrency(10)); err != nil {
		slog.Error("failed to add stage", "error", err)
		os.Exit(1)
	}
	if err = eventSource.Push(&mockStage{index: 1, duration: stageDuration}, eventpipe.OptName("mock 1"), eventpipe.OptConcurrency(10)); err != nil {
		slog.Error("failed to add stage", "error", err)
		os.Exit(1)
	}
	if err = eventSource.Push(&mockStage{index: 2, duration: stageDuration}, eventpipe.OptName("mock 2"), eventpipe.OptConcurrency(10)); err != nil {
		slog.Error("failed to add stage", "error", err)
		os.Exit(1)
	}
	if err = eventSource.Push(&mockStage{index: 3, duration: stageDuration, remove: true}, eventpipe.OptName("mock 3"), eventpipe.OptConcurrency(10)); err != nil {
		slog.Error("failed to add stage", "error", err)
		os.Exit(1)
	}

	slog.Info("starting event source")

	if computeStats {
		stats, err := eventSource.RunAndMeasure()
		if err != nil {
			slog.Error("failed to run event source", "error", err)
			os.Exit(1)
		}
		fmt.Printf("stats: %s\n", stats.String())
	} else {
		if err = eventSource.Run(); err != nil {
			slog.Error("failed to run event source", "error", err)
			os.Exit(1)
		}
	}

	slog.Info("shutting down")
}
