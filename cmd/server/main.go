package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Maxim-Ba/information-keeper/internal/server"
)

func main() {
	//TODO get config

	grpcServer := server.NewGRPCServer()

	go func() {
		if err := grpcServer.Start(":50051"); err != nil {
			slog.Error(err.Error())
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down gRPC server...")

	grpcServer.Stop()
	log.Println("gRPC server stopped")

}
