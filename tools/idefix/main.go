package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"runtime/debug"

	idf "github.com/nayarsystems/idefix-go"
	"github.com/spf13/cobra"
)

var rootctx context.Context
var cancel context.CancelFunc
var version string = getVersion()

// https://tip.golang.org/doc/go1.18#debug_buildinfo
// https://blog.carlana.net/post/2023/golang-git-hash-how-to/
func getVersion() (v string) {
	v = "unknown"
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	dirty := false
	for _, kv := range info.Settings {
		switch kv.Key {
		case "vcs.revision":
			v = kv.Value[:8]
		case "vcs.modified":
			dirty = kv.Value == "true"
		}
	}
	if dirty {
		v += "-dirty"
	}
	return
}

var rootCmd = &cobra.Command{
	Use:          "idefix",
	Short:        fmt.Sprintf("idefix multi-tool (version: %s)", version),
	SilenceUsage: true,
}

func handleInterrupts(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt) // SIGINT

	go func() {
		<-sigChan
		cancel()
	}()
}

func main() {
	rootctx, cancel = context.WithCancel(context.Background())
	handleInterrupts(cancel)

	rootCmd.PersistentFlags().StringP("config", "c", "default", "idefix-go config file for connection settings")
	rootCmd.PersistentFlags().Uint("timeout", 30000, "global timeout in milliseconds")
	rootCmd.PersistentFlags().String("session", "", "session ID")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func getConnectedClient() (*idf.Client, error) {
	configName, err := rootCmd.Flags().GetString("config")
	if err != nil {
		return nil, err
	}

	client, err := idf.NewClientFromFile(rootctx, configName)
	if err != nil {
		return nil, err
	}

	sessionID, err := rootCmd.Flags().GetString("session")
	if err == nil {
		client.SetSessionID(sessionID)
	}

	err = client.Connect()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func getTimeout(cmd *cobra.Command) time.Duration {
	timeout := time.Second * 10
	ptimeout, err := cmd.Flags().GetUint("timeout")
	if err == nil {
		timeout = time.Duration(ptimeout * 1000000)
	}
	return timeout
}
