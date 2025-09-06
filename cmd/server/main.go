package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	config "github.com/Maxim-Ba/information-keeper/config/server"
	"github.com/Maxim-Ba/information-keeper/internal/server"
)

func main() {
	cfg:= config.NewConfig()

	grpcServer := server.NewGRPCServer()

	go func() {
			slog.Info("gRPC server running on " + cfg.GetConfig().ServerHost + ":" + cfg.GetConfig().ServerPort)

		if err := grpcServer.Start(cfg.GetConfig().ServerHost + ":" + cfg.GetConfig().ServerPort); err != nil {
			slog.Error(err.Error())
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	slog.Info("Shutting down gRPC server...")

	grpcServer.Stop()
	slog.Info("gRPC server stopped")

}
