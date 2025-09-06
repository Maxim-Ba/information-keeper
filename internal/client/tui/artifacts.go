package tui

import (
	"context"
	"fmt"

	"github.com/Maxim-Ba/information-keeper/pkg/proto"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func newArtifactsModel() artifactsModel {
	items := []list.Item{}
	
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "📦 Мои артефакты"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	paginationStyle := lipgloss.Style{} //TODO add pagination style
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return artifactsModel{
		list:    l,
		loading: true,
	}
}

func (m artifactsModel) Init() tea.Cmd {
	return m.loadArtifacts
}

func (m artifactsModel) Update(msg tea.Msg) (artifactsModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case artifactsLoadedMsg:
		m.loading = false
		m.artifacts = msg.artifacts
		items := make([]list.Item, len(msg.artifacts))
		for i, artifact := range msg.artifacts {
			items[i] = artifactItem{artifact: artifact}
		}
		m.list.SetItems(items)
	case tea.KeyMsg:
		if msg.String() == "esc" {
			return m, func() tea.Msg { return navigateToMsg{state: mainMenuState} }
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m artifactsModel) View() string {
	if m.loading {
		return "Загрузка артефактов..."
	}
	return m.list.View()
}

func (m artifactsModel) loadArtifacts() tea.Msg {
	ctx := context.Background()
	resp, err := globalClient.ListArtifacts(ctx, globalClient.TokenManager.GetAccessToken(), 1, 50, proto.ArtifactTypeEnum_UNKNOWN)
	if err != nil {
		return errorMsg(fmt.Errorf("Ошибка загрузки артефактов: %v", err))
	}
	if resp.Error != "" {
		return errorMsg(fmt.Errorf("Ошибка: %s", resp.Error))
	}
	return artifactsLoadedMsg{artifacts: resp.Artifacts}
}
