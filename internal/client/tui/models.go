package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Maxim-Ba/information-keeper/internal/client"
	"github.com/Maxim-Ba/information-keeper/pkg/logger"
	"github.com/Maxim-Ba/information-keeper/pkg/proto"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState int

const (
	authState sessionState = iota
	mainMenuState
	artifactsState
	createArtifactState
	settingsState
	artifactDetailState
	viewArtifactDetailsState
)

type model struct {
	state               sessionState
	auth                authModel
	mainMenu            mainMenuModel
	artifacts           artifactsModel
	createArtifact      createArtifactModel
	settings            settingsModel
	artifactDetail      artifactDetailModel
	viewArtifactDetails viewArtifactDetailsModel
	client              *client.GRPCClient
	width               int
	height              int
	err                 error
	successMessage      string
}

type mainMenuModel struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

type artifactsModel struct {
	list      list.Model
	artifacts []*proto.Artifact
	loading   bool
	width     int
	height    int
}

type createArtifactModel struct {
	screen createArtifactScreen
}

type settingsModel struct {
	choices []string
	cursor  int
}

type artifactDetailModel struct {
	artifact *proto.Artifact
	choices  []string
	cursor   int
	width    int
	height   int
}

func newArtifactDetailModel(artifact *proto.Artifact) artifactDetailModel {
	return artifactDetailModel{
		artifact: artifact,
		choices: []string{
			"👀 Просмотреть детали",
			"✏️ Редактировать",
			"🗑️ Удалить",
			"↩️ Назад к списку",
		},
		width:  80,
		height: 20,
	}
}

func (m artifactDetailModel) Init() tea.Cmd {
	return nil
}

func (m artifactDetailModel) Update(msg tea.Msg) (artifactDetailModel, tea.Cmd) {
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
		case "esc":
			return m, func() tea.Msg { return navigateToMsg{state: artifactsState} }
		}
	}
	return m, nil
}

func (m artifactDetailModel) handleSelection() tea.Cmd {
	return func() tea.Msg {
		switch m.cursor {
		case 0: // Просмотреть детали
			return viewArtifactDetailsMsg{artifact: m.artifact}
		case 1: // Редактировать
			return editArtifactMsg{artifact: m.artifact}
		case 2: // Удалить
			return deleteArtifactMsg{artifact: m.artifact}
		case 3: // Назад
			return navigateToMsg{state: artifactsState}
		}
		return nil
	}
}

func (m artifactDetailModel) View() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("📋 Артефакт: %s\n\n", m.artifact.MetaInfo))

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = "▶"
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, choice))
	}

	b.WriteString("\n↑/↓ - навигация • Enter - выбор • Esc - назад\n")

	content := b.String()
	style := lipgloss.NewStyle().
		Width(m.width - 4).
		MaxWidth(80).
		Height(m.height - 4).
		Align(lipgloss.Left)

	return style.Render(content)
}

func InitialModel(grpcClient *client.GRPCClient) model {
	return model{
		state:  authState,
		auth:   newAuthModel(),
		client: grpcClient,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case viewArtifactDetailsMsg:
		logger.Info("🔄 Переход к просмотру деталей артефакта",
			"artifactID", msg.artifact.Id,
			"type", msg.artifact.Type)
		m.state = viewArtifactDetailsState
		m.viewArtifactDetails = newViewArtifactDetailsModel(msg.artifact)
		initCmd := m.viewArtifactDetails.Init()
		return m, initCmd
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		switch m.state {
		case artifactsState:
			listHeight := msg.Height - 4
			listWidth := msg.Width - 2
			logger.Info(fmt.Sprintf("Установка размера списка: %dx%d", listWidth, listHeight))
			m.artifacts.list.SetSize(listWidth, listHeight)
			m.artifacts.width = listWidth
			m.artifacts.height = listHeight
		case createArtifactState:
			// Передаем размеры экрану создания артефакта
			m.createArtifact, cmd = m.createArtifact.Update(msg)
			// Если есть команда от создания артефакта, выполняем ее
			if cmd != nil {
				return m, cmd
			}
		case artifactDetailState:
			// Явно передаем размеры в модель деталей
			m.artifactDetail.width = msg.Width
			m.artifactDetail.height = msg.Height
		case viewArtifactDetailsState:
			m.viewArtifactDetails.width = msg.Width
			m.viewArtifactDetails.height = msg.Height
		}
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.logout()
			return m, tea.Quit
		}

	case successMsg:
		m.successMessage = string(msg)
		// Автоматически очищаем сообщение через 3 секунды
		return m, tea.Tick(3*time.Second, func(time.Time) tea.Msg {
			return clearMessageMsg{}
		})
	case clearMessageMsg:
		m.successMessage = ""
		m.err = nil
	case loginSuccessMsg:
		m.state = mainMenuState
		m.mainMenu = newMainMenuModel()
		return m, nil
	case logoutMsg:
		// После успешного logout возвращаемся к состоянию авторизации
		m.state = authState
		m.auth = newAuthModel()
		m.logout()
		m.successMessage = "Вы успешно вышли из системы"
		return m, tea.Tick(3*time.Second, func(time.Time) tea.Msg {
			return clearMessageMsg{}
		})
	case artifactContentErrorMsg:
		if m.state == viewArtifactDetailsState {
			m.viewArtifactDetails.loading = false
			m.viewArtifactDetails.error = msg.error
		}
	case navigateToMsg:
		// Обрабатываем навигацию между состояниями
		m.state = msg.state

		// Инициализируем соответствующие модели при переходе
		switch msg.state {
		case artifactsState:
			m.artifacts = newArtifactsModel()
			return m, m.artifacts.loadArtifacts
		case createArtifactState:
			m.createArtifact = newCreateArtifactModel()
		case settingsState:
			m.settings = newSettingsModel()
		case mainMenuState:
			m.mainMenu = newMainMenuModel()
		}
		return m, nil
	case artifactSelectedMsg:
		// Переход к деталям артефакта
		m.state = artifactDetailState
		m.artifactDetail = newArtifactDetailModel(msg.artifact)
		return m, nil
	case deleteArtifactMsg:
		// Удаление артефакта
		return m, m.deleteArtifact(msg.artifact)

	// ДЛЯ ОБРАБОТКИ РЕЗУЛЬТАТА СОЗДАНИЯ
	case artifactCreateResultMsg:
		logger.Info("Обработка результата создания артефакта", "success", msg.success)
		if msg.success {
			m.successMessage = msg.message
			// Возвращаемся в главное меню после успешного создания
			m.state = mainMenuState
			m.mainMenu = newMainMenuModel()
			return m, tea.Tick(3*time.Second, func(time.Time) tea.Msg {
				return clearMessageMsg{}
			})
		} else {
			// Устанавливаем ошибку и сбрасываем флаг submitting
			m.err = msg.error
			if m.state == createArtifactState {
				m.createArtifact.screen.submitting = false
			}
			// Автоматически очищаем ошибку через 5 секунд
			return m, tea.Tick(5*time.Second, func(time.Time) tea.Msg {
				return clearMessageMsg{}
			})
		}
	}

	switch m.state {
	case authState:
		m.auth, cmd = m.auth.Update(msg)
	case mainMenuState:
		m.mainMenu, cmd = m.mainMenu.Update(msg)
		// Обработка выбора в главном меню
		if cmd != nil {
			return m, cmd
		}
	case artifactsState:
		logger.Info("Обновление списка артефактов")
		m.artifacts, cmd = m.artifacts.Update(msg)
	case createArtifactState:
		logger.Info("Создание артефакта")
		m.createArtifact, cmd = m.createArtifact.Update(msg)
	case settingsState:
		m.settings, cmd = m.settings.Update(msg)
	case artifactDetailState:
		m.artifactDetail, cmd = m.artifactDetail.Update(msg)
	case viewArtifactDetailsState:
		m.viewArtifactDetails, cmd = m.viewArtifactDetails.Update(msg)

	}

	return m, cmd
}

func (m model) View() string {

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Ошибка: %v", m.err))
	}

	if m.successMessage != "" {
		return successStyle.Render(m.successMessage)
	}

	var view string
	switch m.state {
	case authState:
		view = m.auth.View()
	case mainMenuState:
		view = m.mainMenu.View()
	case artifactsState:
		view = m.artifacts.View()
	case createArtifactState:
		view = m.createArtifact.View()
	case settingsState:
		view = m.settings.View()
	case artifactDetailState:
		view = m.artifactDetail.View()
	case viewArtifactDetailsState:
		view = m.viewArtifactDetails.View()
	}

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		view,
	)
}

func (m model) deleteArtifact(artifact *proto.Artifact) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		_, err := m.client.DeleteArtifact(ctx, artifact.Id)
		if err != nil {
			return errorMsg(fmt.Errorf("ошибка удаления артефакта: %v", err))
		}

		return successMsg("Артефакт успешно удален!")
	}
}

func (m model) logout() tea.Msg {
	logger.Info("Выполнение Logout")

	ctx := context.Background()

	resp, err := m.client.Logout(ctx)
	if err != nil {
		logger.Error("Ошибка при logout", "error", err.Error())
		return errorMsg(fmt.Errorf("ошибка при выходе: %v", err))
	}
	if resp != nil && resp.Error != "" {
		logger.Error("Ошибка от сервера при logout", "error", resp.Error)
		return errorMsg(fmt.Errorf("ошибка сервера: %s", resp.Error))
	}

	// Очищаем токены в любом случае
	m.client.TokenManager.SetTokens("", "")
	logger.Info("Logout выполнен успешно, токены очищены")

	return logoutMsg{}
}

// Сообщения для передачи между компонентами
type errorMsg error
type successMsg string
type clearMessageMsg struct{}
type loginSuccessMsg struct {
	accessToken  string
	refreshToken string
}
type logoutMsg struct{}

type artifactsLoadedMsg struct {
	artifacts []*proto.Artifact
}

type artifactSelectedMsg struct {
	artifact *proto.Artifact
}

type viewArtifactDetailsMsg struct {
	artifact *proto.Artifact
}

type editArtifactMsg struct {
	artifact *proto.Artifact
}

type deleteArtifactMsg struct {
	artifact *proto.Artifact
}

type artifactCreateResultMsg struct {
	success bool
	message string
	error   error
}
