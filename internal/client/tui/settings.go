package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func newSettingsModel() settingsModel {
	return settingsModel{
		choices: []string{
			"🔐 Сменить пароль",
			"📧 Настройки email",
			"↩ Назад",
		},
	}
}

func (m settingsModel) Init() tea.Cmd {
	return nil
}

func (m settingsModel) Update(msg tea.Msg) (settingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case UP, "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case DOWN, "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case ENTER, " ":
			return m, m.handleSelection()
		case ESC:
			return m, func() tea.Msg { return navigateToMsg{state: mainMenuState} }
		}
	}
	return m, nil
}

func (m settingsModel) handleSelection() tea.Cmd {
	return func() tea.Msg {
		switch m.cursor {
		case 0: // Сменить пароль
			// TODO: Реализовать смену пароля
			return successMsg("Функция смены пароля в разработке")
		case 1: // Настройки email
			// TODO: Реализовать настройки email
			return successMsg("Функция настроек email в разработке")
		case 2: // Назад
			return navigateToMsg{state: mainMenuState}
		}
		return nil
	}
}

func (m settingsModel) View() string {
	var b strings.Builder
	b.WriteString("⚙️ Настройки\n\n")

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ARROW
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, choice))
	}

	b.WriteString("\n↑/↓ - навигация • Enter - выбор • Esc - назад\n")
	return b.String()
}
