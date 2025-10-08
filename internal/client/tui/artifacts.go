package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/Maxim-Ba/information-keeper/pkg/logger"
	"github.com/Maxim-Ba/information-keeper/pkg/proto"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Элемент списка для артефактов
type artifactItem struct {
	artifact *proto.Artifact
}

func (i artifactItem) Title() string {
	if i.artifact == nil {
		return "❌ Нет артефакта"
	}
	return i.artifact.MetaInfo
}

func (i artifactItem) Description() string {
	if i.artifact == nil {
		return "❌ Нет данных"
	}
	updatedTime := time.Unix(i.artifact.UpdatedAt, 0).Format("02.01.2006 15:04")
	typeName := getArtifactTypeName(i.artifact.Type)

	return fmt.Sprintf("%s • Обновлен: %s", typeName, updatedTime)
}

func (i artifactItem) FilterValue() string {
	if i.artifact == nil {
		return ""
	}
	return i.artifact.MetaInfo
}

// Добавляем вспомогательную функцию для получения читаемого названия типа
func getArtifactTypeName(artifactType proto.ArtifactTypeEnum) string {
	switch artifactType {
	case proto.ArtifactTypeEnum_TEXT:
		return "📝 Текст"
	case proto.ArtifactTypeEnum_LOGIN_PASSWORD:
		return "🔐 Логин/Пароль"
	case proto.ArtifactTypeEnum_BANK_CARD:
		return "💳 Банковская карта"
	case proto.ArtifactTypeEnum_BINARY:
		return "📎 Файл"
	default:
		return "❓ Неизвестный"
	}
}

func newArtifactsModel() artifactsModel {
	items := []list.Item{}

	delegate := list.NewDefaultDelegate()

	l := list.New(items, delegate, 0, 0)
	l.Title = "📦 Мои артефакты"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.HelpStyle = helpStyle

	l.SetSize(80, 20)

	return artifactsModel{
		list:    l,
		loading: true,
		width:   80,
		height:  20,
	}
}

func (m artifactsModel) Init() tea.Cmd {
	logger.Info("Инициализация модели артефактов")
	return m.loadArtifacts
}

func (m artifactsModel) Update(msg tea.Msg) (artifactsModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case artifactsLoadedMsg:
		m.loading = false
		m.artifacts = msg.artifacts
		logger.Info(fmt.Sprintf("Загружено артефактов: %d", len(msg.artifacts)))

		items := make([]list.Item, len(msg.artifacts))
		for i, artifact := range msg.artifacts {
			items[i] = artifactItem{artifact: artifact}
		}

		m.list.SetItems(items)

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

	view := m.list.View()

	return view
}

func (m artifactsModel) loadArtifacts() tea.Msg {
	ctx := context.Background()
	resp, err := globalClient.ListArtifacts(ctx, 1, 50, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("Ошибка загрузки артефактов: %v", err))
		return errorMsg(fmt.Errorf("ошибка загрузки артефактов: %v", err))
	}
	if resp.Error != "" {
		logger.Error(fmt.Sprintf("Ошибка от сервера: %s", resp.Error))
		return errorMsg(fmt.Errorf("ошибка: %s", resp.Error))
	}
	logger.Info(fmt.Sprintf("Успешно загружено %d артефактов", len(resp.Artifacts)))
	return artifactsLoadedMsg{artifacts: resp.Artifacts}
}
