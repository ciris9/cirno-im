package service

import (
	"cirno-im/logger"
	"cirno-im/naming"
	"cirno-im/naming/consul"
	"cirno-im/services/service/conf"
	"cirno-im/services/service/database"
	"cirno-im/services/service/handler"
	"cirno-im/wire"
	"context"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/spf13/cobra"
	"hash/crc32"
)

type ServerStartOptions struct {
	config string
}

func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOptions{}
	cmd := &cobra.Command{
		Use:   "royal",
		Short: "start a rpc services",
		RunE:  func(cmd *cobra.Command, args []string) error { return RunServerStart(ctx, opts, version) },
	}
	cmd.PersistentFlags().StringVarP(&opts.config, "config", "c", "conf.yaml", "config file")
	return cmd
}

func RunServerStart(ctx context.Context, opts *ServerStartOptions, version string) error {
	config, err := conf.Init(opts.config)
	if err != nil {
		return err
	}
	err = logger.Init(logger.Setting{
		Filename: "./data/royal.log",
		Level:    config.LogLevel,
	})
	if err != nil {
		return err
	}

	db, err := database.InitMysqlDB(config.BaseDb)
	if err != nil {
		return err
	}
	_ = db.AutoMigrate(&database.Group{}, &database.GroupMember{})

	messageDb, err := database.InitMysqlDB(config.MessageDb)
	if err != nil {
		return err
	}
	_ = messageDb.AutoMigrate(&database.MessageIndex{}, &database.MessageContent{})

	if config.NodeID == 0 {
		config.NodeID = int64(HashCode(config.ServiceID))
	}
	IdGen, err := database.NewIDGenerator(config.NodeID)
	if err != nil {
		return err
	}

	rdb, err := database.InitRedis(config.RedisAddrs, "")
	if err != nil {
		return err
	}

	ns, err := consul.NewNaming(config.ConsulURL)
	if err != nil {
		return err
	}

	if err = ns.Register(&naming.DefaultService{
		Name:     wire.SNService, // service name
		Address:  config.PublicAddress,
		Port:     config.PublicPort,
		Protocol: "http",
		Tags:     config.Tags,
		Meta: map[string]string{
			consul.KeyHealthURL: fmt.Sprintf("http://%s:%d/health", config.PublicAddress, config.PublicPort),
		},
	}); err != nil {
		return err
	}
	defer func() {
		err = ns.Deregister(config.ServiceID)
		if err != nil {
			logger.Errorf("deregister service %s err: %v", config.ServiceID, err)
		}
	}()

	serviceHandler := handler.ServiceHandler{
		BaseDb:    db,
		MessageDb: messageDb,
		Cache:     rdb,
		IdGen:     IdGen,
	}
	ac := conf.MakeAccessLog()
	defer ac.Close()

	app := newApp(&serviceHandler)
	app.UseRouter(ac.Handler)
	app.UseRouter(setAllowedResponses)
	return app.Listen(config.Listen, iris.WithOptimizations)
}

func newApp(serviceHandler *handler.ServiceHandler) *iris.Application {
	app := iris.Default()
	app.Get("/health", func(ctx iris.Context) {
		_, _ = ctx.WriteString("ok")
	})
	messageApi := app.Party("/api/:app/message")
	{
		messageApi.Post("/user", serviceHandler.InsertUserMessage)
		messageApi.Post("/group", serviceHandler.InsertGroupMessage)
		messageApi.Post("/ack", serviceHandler.MessageAck)
	}
	groupApi := app.Party("/api/:app/group")
	{
		groupApi.Get("/:id", serviceHandler.GroupGet)
		groupApi.Post("", serviceHandler.GroupCreate)
		groupApi.Post("/member", serviceHandler.GroupJoin)
		groupApi.Delete("/member", serviceHandler.GroupQuit)
		groupApi.Get("/members/:id", serviceHandler.GroupMembers)
	}

	offlineApi := app.Party("/api/:app/offline")
	{
		offlineApi.Use(iris.Compression)
		offlineApi.Post("/index", serviceHandler.GetOfflineMessageIndex)
		offlineApi.Post("/content", serviceHandler.GetOfflineMessageContent)
	}
	return app
}

func setAllowedResponses(ctx iris.Context) {
	ctx.Negotiation().JSON().Protobuf().MsgPack()
	ctx.Negotiation().Accept.JSON()
	ctx.Next()
}

func HashCode(id string) uint32 {
	hash32 := crc32.NewIEEE()
	_, _ = hash32.Write([]byte(id))
	return hash32.Sum32() % 1000
}
