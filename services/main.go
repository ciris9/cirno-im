package main

import (
	"context"
	"flag"

	"cirno-im/logger"
	"cirno-im/services/gateway"
	"cirno-im/services/server"
	"github.com/spf13/cobra"
)

const version = "v1"

func main() {
	flag.Parse()

	root := &cobra.Command{
		Use:     "cim",
		Version: version,
		Short:   "cirno-im service",
	}
	ctx := context.Background()

	root.AddCommand(gateway.NewServerStartCMD(ctx, version))
	root.AddCommand(server.NewServerStartCMD(ctx, version))

	if err := root.Execute(); err != nil {
		logger.WithError(err).Fatal("Could not run command")
	}
}
