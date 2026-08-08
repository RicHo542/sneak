package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/richo542/sneak/internal/cli"
)

// set via -ldflags at build time
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	root := cli.NewRootCmd(cli.BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	})

	root.SetContext(ctx)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
