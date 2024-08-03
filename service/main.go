package main

import (
	"context"
	"flag"

	"cirno-im/logger"
	"cirno-im/service/gateway"
	"cirno-im/service/server"
	"github.com/spf13/cobra"
)

const version = "v1"

func main() {
	flag.Parse()

	root := &cobra.Command{
		Use:     "kim",
		Version: version,
		Short:   "King IM Cloud",
	}
	ctx := context.Background()

	root.AddCommand(gateway.NewServerStartCMD(ctx, version))
	root.AddCommand(server.NewServerStartCMD(ctx, version))

	if err := root.Execute(); err != nil {
		logger.WithError(err).Fatal("Could not run command")
	}
}
