package main

import (
	"context"
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
	"github.com/Maxim-Ba/information-keeper/pkg/utils"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}
	lgCfg := logger.DefaultConfig()
	lgCfg.AddSource = false
	if err := logger.InitLogger(lgCfg); err != nil {
		panic(err)
	}
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
	if err != nil {
		panic(err.Error())
	}
	defer func() {
		conn.Close()
	}()

	tokenRepo := repository.NewTokenRepository(conn)
	userRepo := repository.NewUserRepository(conn, cfg, &utils.PasswordManager{})
	artifactRepo := repository.NewArtifactRepository(s3client, conn)
	tokenService := services.NewTokenService(cfg, userRepo, tokenRepo)
	authService := services.New(userRepo, tokenService, cfg)
	syncManager := services.NewSyncManager(5, time.Duration(5)) // TODO : change to config

	artifactService := services.NewArtifactService(artifactRepo, syncManager)
	grpcServer := server.NewGRPCServer(authService, artifactService, tokenService, syncManager)
	go func() {
		logger.Info("gRPC server running on " + cfg.GetConfig().ServerHost + ":" + cfg.GetConfig().ServerPort)

		if err := grpcServer.Start(cfg.GetConfig().ServerHost + ":" + cfg.GetConfig().ServerPort); err != nil {
			logger.Error(err.Error())
		}
	}()

	<-ctx.Done()
	logger.Info("Received shutdown signal, initiating graceful shutdown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		grpcServer.Stop()
		cancel()
	}()

	<-shutdownCtx.Done()

	if shutdownCtx.Err() == context.DeadlineExceeded {
		logger.Warn("Graceful shutdown timed out, forcing exit")
	} else {
		logger.Info("gRPC server stopped gracefully")
	}

}
