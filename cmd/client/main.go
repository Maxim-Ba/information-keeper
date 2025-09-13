package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	config "github.com/Maxim-Ba/information-keeper/config/client"
	"github.com/Maxim-Ba/information-keeper/internal/client"
	"github.com/Maxim-Ba/information-keeper/internal/client/tui"
	"github.com/Maxim-Ba/information-keeper/pkg/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	cfg := config.NewConfig()
	lconfig := logger.DefaultConfig()
	lconfig.FilePath = "myapp.log" // TODO : change to config
	lconfig.Level = logger.LevelDebug

	if err := logger.InitLogger(lconfig); err != nil {
		panic(err)
	}

	clnt, err := client.NewGRPCClient(cfg.GetConfig().ServerHost + ":" + cfg.GetConfig().ServerPort)
	if err != nil {
		panic(err)
	}
	defer clnt.Close()

	healthCtx, healthCancel := context.WithTimeout(ctx, 5*time.Second)
	defer healthCancel()

	if err := clnt.HealthCheck(healthCtx); err != nil {
		panic(fmt.Sprintf("Не удалось подключиться к серверу: %v", err))
	}

	if err := tui.StartTUIWithContext(ctx, clnt); err != nil {
		panic(fmt.Sprintf("Ошибка запуска TUI: %v", err))
	}

}
