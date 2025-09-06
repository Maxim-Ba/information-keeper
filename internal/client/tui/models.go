package tui

import (
	"fmt"
	"time"

	"github.com/Maxim-Ba/information-keeper/internal/client"
	"github.com/Maxim-Ba/information-keeper/pkg/proto"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
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
)

// Основная модель
type model struct {
	state           sessionState
	auth            authModel
	mainMenu        mainMenuModel
	artifacts       artifactsModel
	createArtifact  createArtifactModel
	settings        settingsModel
	client          *client.GRPCClient
	width           int
	height          int
	err             error
	successMessage  string
}

// Модель главного меню
type mainMenuModel struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

// Модель списка артефактов
type artifactsModel struct {
	list     list.Model
	artifacts []*proto.Artifact
	loading  bool
}

// Модель создания артефакта
type createArtifactModel struct {
	inputs     []textinput.Model
	focusIndex int
	artifactType int
	submitting bool
}

// Модель настроек
type settingsModel struct {
	choices []string
	cursor  int
}

// Элемент списка для артефактов
type artifactItem struct {
	artifact *proto.Artifact
}

func (i artifactItem) Title() string       { return i.artifact.MetaInfo }
func (i artifactItem) Description() string { 
	return fmt.Sprintf("Type: %s, Created: %d", i.artifact.Type.String(), i.artifact.CreatedAt)
}
func (i artifactItem) FilterValue() string { return i.artifact.MetaInfo }

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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Передаем размеры дочерним компонентам
		switch m.state {
		case artifactsState:
			m.artifacts.list.SetSize(msg.Width, msg.Height)
		}
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	case errorMsg:
		m.err = msg
		return m, nil
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
	}

	// Делегируем обновление текущему состоянию
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
		m.artifacts.list, cmd = m.artifacts.list.Update(msg)
	case createArtifactState:
		m.createArtifact, cmd = m.createArtifact.Update(msg)
	case settingsState:
		m.settings, cmd = m.settings.Update(msg)
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
	}

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		view,
	)
}

// Сообщения для передачи между компонентами
type errorMsg error
type successMsg string
type clearMessageMsg struct{}
type loginSuccessMsg struct {
	accessToken  string
	refreshToken string
}
type artifactsLoadedMsg struct {
	artifacts []*proto.Artifact
}
