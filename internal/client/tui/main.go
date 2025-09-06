package tui

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/Maxim-Ba/information-keeper/internal/client"
	tea "github.com/charmbracelet/bubbletea"
)

var globalClient *client.GRPCClient

func StartTUI(grpcClient *client.GRPCClient) {
	globalClient = grpcClient

	if err := grpcClient.HealthCheck(context.TODO()); err != nil {
		panic(fmt.Sprintf("Не удалось подключиться к серверу: %v", err))
	}

	p := tea.NewProgram(InitialModel(grpcClient), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		slog.Info(fmt.Sprintf("Ошибка запуска TUI: %v", err)) 
		os.Exit(1)
	}
}
