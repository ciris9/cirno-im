package main

import (
	"cirno-im/services/router"
	"cirno-im/services/service"
	"cirno-im/trace"
	"context"
	"flag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"log"

	"cirno-im/logger"
	"cirno-im/services/gateway"
	"cirno-im/services/server"
	"github.com/spf13/cobra"
)

const version = "v1"

const jaegerTraceProviderAddress = ""

func main() {
	flag.Parse()

	tp, tpErr := trace.JaegerTraceProvider(jaegerTraceProviderAddress)
	if tpErr != nil {
		log.Fatal(tpErr)
	}
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	root := &cobra.Command{
		Use:     "cim",
		Version: version,
		Short:   "cirno-im service",
	}
	ctx := context.Background()

	root.AddCommand(gateway.NewServerStartCMD(ctx, version))
	root.AddCommand(server.NewServerStartCMD(ctx, version))
	root.AddCommand(service.NewServerStartCmd(ctx, version))
	root.AddCommand(router.NewServerStartCmd(ctx, version))

	if err := root.Execute(); err != nil {
		logger.WithError(err).Fatal("Could not run command")
	}
}
