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
		// Обновляем размер списка после загрузки данных
		if m.width > 0 && m.height > 0 {
			m.list.SetSize(m.width, m.height)
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return navigateToMsg{state: mainMenuState} }
		case "enter":
			if !m.loading && len(m.artifacts) > 0 {
				selectedItem := m.list.SelectedItem()
				if item, ok := selectedItem.(artifactItem); ok {
					return m, func() tea.Msg { 
						return artifactSelectedMsg{artifact: item.artifact} 
					}
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height)
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m artifactsModel) View() string {
	if m.loading {
		return "Загрузка артефактов..."
	}
	
	if len(m.artifacts) == 0 {
		return "📭 У вас пока нет артефактов\n\nEsc - назад"
	}
	
	return m.list.View()
}

func (m artifactsModel) loadArtifacts() tea.Msg {
	ctx := context.Background()
	resp, err := globalClient.ListArtifacts(ctx,  1, 50, proto.ArtifactTypeEnum_UNKNOWN)
	if err != nil {
		return errorMsg(fmt.Errorf("Ошибка загрузки артефактов: %v", err))
	}
	if resp.Error != "" {
		return errorMsg(fmt.Errorf("Ошибка: %s", resp.Error))
	}
	return artifactsLoadedMsg{artifacts: resp.Artifacts}
}
