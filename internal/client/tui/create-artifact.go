package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/Maxim-Ba/information-keeper/pkg/logger"
	"github.com/Maxim-Ba/information-keeper/pkg/proto"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newCreateArtifactModel() createArtifactModel {
	m := createArtifactModel{
		inputs: make([]textinput.Model, 3),
	}

	// Поле для мета-информации
	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "Мета-информация"
	m.inputs[0].Focus()
	m.inputs[0].Width = 40

	// Поле для ссылки
	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = "Ссылка (опционально)"
	m.inputs[1].Width = 40

	// Поле для времени жизни
	m.inputs[2] = textinput.New()
	m.inputs[2].Placeholder = "Время жизни (в часах, опционально)"
	m.inputs[2].Width = 40

	return m
}

func (m createArtifactModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m createArtifactModel) Update(msg tea.Msg) (createArtifactModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return navigateToMsg{state: mainMenuState} }
		case "tab", "shift+tab", "up", "down":
			// Навигация по полям
			s := msg.String()

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs)-1 {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i < len(m.inputs); i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}
			return m, tea.Batch(cmds...)
		case "enter":
			if m.submitting {
				return m, nil
			}
			// Проверяем обязательное поле
			if strings.TrimSpace(m.inputs[0].Value()) == "" {
				return m, nil
			}
			m.submitting = true
			return m, m.submitArtifact()
		}
	}

	// Обновляем поля ввода
	var cmd tea.Cmd
	for i := range m.inputs {
		m.inputs[i], cmd = m.inputs[i].Update(msg)
	}

	return m, cmd
}

func (m createArtifactModel) submitArtifact() tea.Cmd {
	return func() tea.Msg {
		metaInfo := strings.TrimSpace(m.inputs[0].Value())
		link := strings.TrimSpace(m.inputs[1].Value())
		// expireHours := strings.TrimSpace(m.inputs[2].Value())

		req := &proto.CreateArtifactRequest{
			Type:     proto.ArtifactTypeEnum_TEXT,
			MetaInfo: metaInfo,
			Link:     link,
		}

		// TODO: Парсинг expireHours

		ctx := context.Background()
		resp, err := globalClient.CreateArtifact(ctx, req)
		if err != nil {
			logger.Error("Ошибка создания артефакта err:", err)
			return artifactCreateResultMsg{
				success: false,
				error:   fmt.Errorf("ошибка создания артефакта: %v", err),
			}
		}
		if resp.Error != "" {
			logger.Error("Ошибка создания артефакта resp.Error:", resp.Error)

			return artifactCreateResultMsg{
				success: false,
				error:   fmt.Errorf("ошибка: %s", resp.Error),
			}
		}

		return artifactCreateResultMsg{
			success: true,
			message: "Артефакт успешно создан!",
		}
	}
}

func (m createArtifactModel) View() string {
	var b strings.Builder
	b.WriteString("➕ Создание артефакта\n\n")

	if m.submitting {
		b.WriteString("⏳ Создание...\n\n")
		return b.String()
	}

	fields := []string{
		"Мета-информация:",
		m.inputs[0].View(),
		"\nСсылка:",
		m.inputs[1].View(),
		"\nВремя жизни (часы):",
		m.inputs[2].View(),
	}

	b.WriteString(strings.Join(fields, "\n"))
	b.WriteString("\n\nEnter - создать • Esc - назад • Tab - переключение полей\n")

	// Подсказка про обязательное поле
	if strings.TrimSpace(m.inputs[0].Value()) == "" {
		b.WriteString("\n⚠️  Мета-информация обязательна для заполнения")
	}

	return b.String()
}

type artifactCreateResultMsg struct {
	success bool
	message string
	error   error
}
