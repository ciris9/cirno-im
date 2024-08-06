package gateway

import (
	cim "cirno-im"
	"cirno-im/container"
	"cirno-im/logger"
	"cirno-im/naming"
	"cirno-im/naming/consul"
	"cirno-im/services/gateway/conf"
	"cirno-im/services/gateway/serv"
	"cirno-im/websocket"
	"cirno-im/wire"
	"context"
	"errors"
	"github.com/spf13/cobra"
	"time"
)

type ServerStartOptions struct {
	config   string
	protocol string
}

func NewServerStartCMD(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOptions{}

	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "start a gateway server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.config, "conf", "c", "./gateway/conf.yaml", "conf file")
	cmd.PersistentFlags().StringVarP(&opts.protocol, "protocol", "p", "ws", "protocol of ws or tcp")
	return cmd
}

func RunServerStart(ctx context.Context, opts *ServerStartOptions, version string) error {
	config, err := conf.Init(opts.config)
	if err != nil {
		return err
	}
	err = logger.Init(logger.Setting{
		Level: "trace",
	})
	if err != nil {
		return err
	}
	handler := &serv.Handler{ServiceID: config.ServiceID}

	var srv cim.Server
	service := &naming.DefaultService{
		Id:       config.ServiceID,
		Name:     config.ServiceName,
		Address:  config.PublicAddress,
		Port:     config.PublicPort,
		Protocol: opts.protocol,
		Tags:     config.Tags,
	}
	logger.Debugln(service)
	if opts.protocol == "ws" {
		srv = websocket.NewServer(config.Listen, service)
	}
	if srv == nil {
		return errors.New("services is nil")
	}
	srv.SetReadWait(time.Minute * 2)
	srv.SetAcceptor(handler)
	srv.SetMessageListener(handler)
	srv.SetStateListener(handler)

	err = container.Init(srv, wire.SNChat, wire.SNLogin)
	if err != nil {
		return err
	}
	ns, err := consul.NewNaming(config.ConsulURL)
	if err != nil {
		return err
	}
	container.SetServiceNaming(ns)

	container.SetDialer(serv.NewDialer(config.ServiceID))
	return container.Start()
}
