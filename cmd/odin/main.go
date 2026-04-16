// ODIN AI - Nórdico Local-First AI Ecosystem
// Core CLI entrypoint
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/odin-ai/odin/internal/cli"
	"github.com/odin-ai/odin/pkg/logger"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutting down ODIN...")
		cancel()
	}()

	// Execute CLI
	cmd := cli.NewRootCmd(version, buildTime)
	if err := cmd.ExecuteContext(ctx); err != nil {
		logger.Error("Application error", "error", err)
		os.Exit(1)
	}
}

// Execute runs the CLI - called by tests
func Execute(version, buildTime string) int {
	cmd := cli.NewRootCmd(version, buildTime)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}
