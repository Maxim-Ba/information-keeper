package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	config "github.com/Maxim-Ba/information-keeper/config/server"
	"github.com/Maxim-Ba/information-keeper/internal/server"
	"github.com/Maxim-Ba/information-keeper/internal/server/repository"
	"github.com/Maxim-Ba/information-keeper/internal/server/s3client"
	"github.com/Maxim-Ba/information-keeper/internal/server/services"
	"github.com/Maxim-Ba/information-keeper/pkg/db"
	"github.com/Maxim-Ba/information-keeper/pkg/logger"
)

func main() {
	cfg := config.NewConfig()

	defer logger.CloseLogger()
	s3client, err := s3client.New(s3client.Config{
		Endpoint:        cfg.GetConfig().S3endpoint,
		AccessKeyID:     cfg.GetConfig().S3accessKey,
		SecretAccessKey: cfg.GetConfig().S3secretKey,
		Region:          cfg.GetConfig().S3region,
		Bucket:          cfg.GetConfig().S3bucket,
		UseSSL:          cfg.GetConfig().S3useSSL,
		ForcePathStyle:  cfg.GetConfig().S3forcePathStyle,
	})
	if err != nil {
		panic(err)
	}
	defer s3client.Close()
	conn, err := db.New(cfg.GetConfig().DBConnectionString, cfg.GetConfig().MigrationsPath)

	tokenRepo := repository.NewTokenRepository(conn)
	userRepo := repository.NewUserRepository(conn)
	artifactRepo := repository.NewArtifactRepository(s3client, conn)
	tokenService := services.NewTokenService(cfg, userRepo, tokenRepo)
	authService := services.New(userRepo, tokenService, cfg)
	syncManager := services.NewSyncManager(5, time.Duration(5)) // TODO : change to config

	artifactService := services.NewArtifactService(artifactRepo, syncManager)
	grpcServer := server.NewGRPCServer(authService, artifactService)
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
