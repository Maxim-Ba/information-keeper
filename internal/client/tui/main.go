package tui

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Maxim-Ba/information-keeper/internal/client"
	tea "github.com/charmbracelet/bubbletea"
)

var globalClient *client.GRPCClient

func StartTUIWithContext(ctx context.Context, grpcClient *client.GRPCClient) error {
	globalClient = grpcClient

	if err := grpcClient.HealthCheck(ctx); err != nil {
		return fmt.Errorf("не удалось подключиться к серверу: %v", err)
	}

	p := tea.NewProgram(InitialModel(grpcClient), tea.WithAltScreen())
	
	go func() {
		if _, err := p.Run(); err != nil {
			slog.Error("Ошибка TUI", "error", err)
		}
	}()
<-ctx.Done()
p.Quit()
return nil
}
