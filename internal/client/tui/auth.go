package tui

import (
	"context"
	"fmt"

	"github.com/Maxim-Ba/information-keeper/pkg/logger"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type authModel struct {
	inputs     []textinput.Model
	focusIndex int
	isLogin    bool
	submitting bool
	err        error
	showError  bool
}

func newAuthModel() authModel {
	loginInput := textinput.New()
	loginInput.Placeholder = "Логин"
	loginInput.Focus()
	loginInput.CharLimit = 50
	loginInput.Width = 30

	passwordInput := textinput.New()
	passwordInput.Placeholder = "Пароль"
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '•'
	passwordInput.CharLimit = 50
	passwordInput.Width = 30

	emailInput := textinput.New()
	emailInput.Placeholder = "Email (для регистрации)"
	emailInput.CharLimit = 100
	emailInput.Width = 30

	return authModel{
		inputs:  []textinput.Model{loginInput, passwordInput, emailInput},
		isLogin: true,
	}
}

func (m authModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m authModel) Update(msg tea.Msg) (authModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab, tea.KeyShiftTab, tea.KeyEnter, tea.KeyUp, tea.KeyDown:
			// Обработка навигации
			s := msg.String()

			if s == ENTER {
				if m.submitting {
					return m, nil
				}
				// Сбрасываем ошибку при новой попытке
				m.err = nil
				m.showError = false
				return m, m.submitAuth()
			}

			if s == "tab" || s == SHIFT_TAB || s == UP || s == DOWN {
				// Смена фокуса между полями ввода
				if s == UP || s == SHIFT_TAB {
					m.focusIndex--
				} else {
					m.focusIndex++
				}

				if m.focusIndex > len(m.inputs) {
					m.focusIndex = 0
				} else if m.focusIndex < 0 {
					m.focusIndex = len(m.inputs)
				}

				cmds := make([]tea.Cmd, len(m.inputs))
				for i := 0; i <= len(m.inputs)-1; i++ {
					if i == m.focusIndex {
						cmds[i] = m.inputs[i].Focus()
					} else {
						m.inputs[i].Blur()
					}
				}

				return m, tea.Batch(cmds...)
			}
		case tea.KeyCtrlT:
			// Переключение между логином и регистрацией
			m.isLogin = !m.isLogin
			// Сбрасываем ошибку при переключении режима
			m.err = nil
			m.showError = false
			if m.isLogin && len(m.inputs) > 2 {
				m.inputs = m.inputs[:2]
			} else if !m.isLogin && len(m.inputs) == 2 {
				emailInput := textinput.New()
				emailInput.Placeholder = "Email"
				emailInput.CharLimit = 100
				emailInput.Width = 30
				m.inputs = append(m.inputs, emailInput)
			}
			return m, nil
		case tea.KeyEscape, tea.KeyBackspace:
			// Скрываем ошибку при нажатии Escape или Backspace
			if m.err != nil {
				m.err = nil
				m.showError = false
				return m, nil
			}
		}
	case errorMsg:
		// Получаем ошибку и устанавливаем флаг отображения
		logger.Error("Update showError =")
		m.err = error(msg)
		m.showError = true
		return m, nil
	}

	// Обновляем поля ввода
	var cmd tea.Cmd
	for i := range m.inputs {
		m.inputs[i], cmd = m.inputs[i].Update(msg)
	}

	return m, cmd
}

func (m authModel) submitAuth() tea.Cmd {
	return func() tea.Msg {
		m.submitting = true
		defer func() { m.submitting = false }()

		login := m.inputs[0].Value()
		password := m.inputs[1].Value()

		ctx := context.Background()
		logger.Info("authModel submitAuth", "login", login, "password", password)
		if m.isLogin {
			logger.Info(fmt.Sprintf("m.isLogin = %v", m.isLogin))
			resp, err := globalClient.Login(ctx, login, password)
			logger.Info(fmt.Sprintf("resp = %v", resp))

			if err != nil {
				logger.Error("authModel submitAuth", "error", err.Error())
				return errorMsg(fmt.Errorf("ошибка авторизации: %v", err))
			}
			if resp.Error != "" {
				return errorMsg(fmt.Errorf("ошибка: %s", resp.Error))
			}
			return loginSuccessMsg{
				accessToken:  resp.AccessToken,
				refreshToken: resp.RefreshToken,
			}
		} else {
			email := m.inputs[2].Value()
			resp, err := globalClient.Register(ctx, login, password, email)
			if err != nil {
				return errorMsg(fmt.Errorf("ошибка регистрации: %v", err))
			}
			if resp.Error != "" {
				return errorMsg(fmt.Errorf("ошибка: %s", resp.Error))
			}
			return successMsg("Регистрация успешна! Теперь вы можете войти.")
		}
	}
}

func (m authModel) View() string {
	title := "🔐 Авторизация"
	if !m.isLogin {
		title = "📝 Регистрация"
	}

	var fields []string
	for i, input := range m.inputs {
		if i == 2 && m.isLogin {
			continue
		}
		fieldName := ""
		switch i {
		case 0:
			fieldName = "Логин:"
		case 1:
			fieldName = "Пароль:"
		case 2:
			fieldName = "Email:"
		}
		fields = append(fields, fmt.Sprintf("%s\n%s", fieldName, input.View()))
	}

	actionBtn := "↳ Войти [Enter]"
	if !m.isLogin {
		actionBtn = "↳ Зарегистрироваться [Enter]"
	}

	toggleText := "Нет аккаунта? [Ctrl+T] - Регистрация"
	if !m.isLogin {
		toggleText = "Уже есть аккаунт? [Ctrl+T] - Вход"
	}

	// Добавляем отображение ошибки, если она есть и нужно показывать
	errorSection := ""
	if m.err != nil && m.showError {
		errorSection = fmt.Sprintf("\n\n%s\n%s",
			errorStyle.Render("Ошибка:"),
			errorStyle.Render(m.err.Error()+"\n[Esc] - Скрыть ошибку"),
		)
	}

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s\n%s\n%s\n\n%s",
		titleStyle.Render(title),
		lipgloss.JoinVertical(lipgloss.Left, fields...),
		actionBtn,
		errorSection,
		helpStyle.Render(toggleText),
		helpStyle.Render("Ctrl+C - Выход | Esc - Скрыть ошибку"),
	)
}
