package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"alex-one/subscriber/internal/app/grpchandler"
	"alex-one/subscriber/internal/domain/service"
	proto "alex-one/subscriber/internal/pkg/api"
	"alex-one/subscriber/internal/pkg/config"
	"alex-one/subscriber/internal/pkg/logger"
	"alex-one/subscriber/internal/pkg/subpub"
)

func main() {
	ctx := context.Background()
	configFile := os.Getenv("CONFIG_FILE")
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		panic(err)
	}

	setupLogger(cfg)

	sp := subpub.NewSubPub()
	subPuber := service.NewSubPuberService(sp)
	grpcPublisher := grpchandler.NewGRPCPublisherServer(subPuber)
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	proto.RegisterPubSubServer(grpcServer, grpcPublisher)

	l := getListener(cfg)
	if l == nil {
		panic("listen error")
	}

	go func() {
		if err := grpcServer.Serve(l); err != nil {
			log.Fatalf("Listen grpc could not listen: %v\n", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Infof(ctx, "start gracefull shotdown")
	sp.Close(ctx)
	grpcServer.GracefulStop()
	logger.Infof(ctx, "end gracefull shotdown")
}

func getListener(cfg *config.Config) net.Listener {
	address := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil
	}
	return listener

}

func setupLogger(cfg *config.Config) {
	logCfg := logger.ConfigLogger{
		Level: cfg.Logger.Level,
	}
	zLog := logger.NewZerologLogger(logCfg)
	logger.SetupLogger(zLog)
}
