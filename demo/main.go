package main

import (
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
		Use:     "fim",
		Version: version,
		Short:   "server",
	}
	ctx := context.Background()

	// mock
	root.AddCommand(mock.NewClientCmd(ctx))
	root.AddCommand(mock.NewServerCmd(ctx))

	if err := root.Execute(); err != nil {
		logger.WithError(err).Fatal("Could not run command")
	}
}
