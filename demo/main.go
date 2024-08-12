package main

import (
	"cirno-im/demo/cimbench"
	"cirno-im/demo/echo"
	"context"
	"flag"

	"cirno-im/demo/mock"
	"cirno-im/logger"
	"github.com/spf13/cobra"
)

const version = "v1"

func main() {
	flag.Parse()

	root := &cobra.Command{
		Use:     "kim",
		Version: version,
		Short:   "tools",
	}
	ctx := context.Background()

	// run echo client
	root.AddCommand(echo.NewCmd(ctx))

	// mock
	root.AddCommand(mock.NewClientCmd(ctx))
	root.AddCommand(mock.NewServerCmd(ctx))
	root.AddCommand(cimbench.NewBenchmarkCmd(ctx))

	if err := root.Execute(); err != nil {
		logger.WithError(err).Fatal("Could not run command")
	}
}
