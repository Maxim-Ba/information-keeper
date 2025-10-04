package tui

import (
	"fmt"
	"strings"

	"github.com/Maxim-Ba/information-keeper/pkg/logger"
	tea "github.com/charmbracelet/bubbletea"
)

func newMainMenuModel() mainMenuModel {
	return mainMenuModel{
		choices: []string{
			"📦 Мои артефакты",
			"➕ Создать артефакт",
			"⚙️ Настройки",
			"🚪 Выйти",
		},
		selected: make(map[int]struct{}),
	}
}

func (m mainMenuModel) Init() tea.Cmd {
	return nil
}

func (m mainMenuModel) Update(msg tea.Msg) (mainMenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			return m, m.handleSelection()
		}
	}
	return m, nil
}

func (m mainMenuModel) handleSelection() tea.Cmd {
	return func() tea.Msg {
		switch m.cursor {
		case 0: // Мои артефакты
			return navigateToMsg{state: artifactsState}
		case 1: // Создать артефакт
			return navigateToMsg{state: createArtifactState}
		case 2: // Настройки
			return navigateToMsg{state: settingsState}
		case 3: // Выйти
		logger.Info("Выход из приложения")
			return logoutMsg{}
		}
		return nil
	}
}

func (m mainMenuModel) View() string {
	var b strings.Builder
	b.WriteString("🏠 Главное меню\n\n")

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = "▶"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		b.WriteString(fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice))
	}

	b.WriteString("\n↑/↓ - навигация • Enter - выбор • Ctrl+C - выход\n")
	return b.String()
}

type navigateToMsg struct {
	state sessionState
}
