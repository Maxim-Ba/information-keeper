package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Maxim-Ba/information-keeper/internal/domain"
	"github.com/Maxim-Ba/information-keeper/pkg/logger"
	"github.com/Maxim-Ba/information-keeper/pkg/proto"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)

// Внутренняя структура для экрана создания артефакта
type createArtifactScreen struct {
	state        createArtifactScreenState
	typeSelector list.Model
	inputs       map[string]textinput.Model
	focusIndex   int
	submitting   bool
	artifactType proto.ArtifactTypeEnum
}

type createArtifactScreenState int

const (
	selectTypeScreenState createArtifactScreenState = iota
	fillDataScreenState
)

type artifactTypeItem struct {
	name string
	desc string
	typ  proto.ArtifactTypeEnum
}

func (i artifactTypeItem) Title() string       { return i.name }
func (i artifactTypeItem) Description() string { return i.desc }
func (i artifactTypeItem) FilterValue() string { return i.name }

// Структуры данных для разных типов артефактов
type TextData domain.TextData

//  struct {
// 	Content string `json:"content"`
// 	Title   string `json:"title,omitempty"`
// }

type LoginPasswordData domain.LoginPasswordData

// struct {
// 	Login    string `json:"login"`
// 	Password string `json:"password"`
// 	Site     string `json:"site,omitempty"`
// 	Notes    string `json:"notes,omitempty"`
// }

type BankCardData domain.BankCardData

// struct {
// 	Number     string `json:"number"`
// 	Holder     string `json:"holder"`
// 	ExpiryDate string `json:"expiry_date"`
// 	CVV        string `json:"cvv,omitempty"`
// }

type BinaryData struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Description string `json:"description,omitempty"`
	Data        []byte `json:"data"`
}

// Методы для TextData
func (d *TextData) ToBinary() ([]byte, error) {
	return json.Marshal(d)
}

func (d *TextData) GetMetaInfo() string {
	if d.Title != "" {
		return d.Title
	}
	if len(d.Content) > 50 {
		return d.Content[:50] + "..."
	}
	return d.Content
}

// Методы для LoginPasswordData
func (d *LoginPasswordData) ToBinary() ([]byte, error) {
	return json.Marshal(d)
}

func (d *LoginPasswordData) GetMetaInfo() string {
	if d.Site != "" {
		return fmt.Sprintf("Логин для %s", d.Site)
	}
	return "Логин и пароль"
}

// Методы для BankCardData
func (d *BankCardData) ToBinary() ([]byte, error) {
	return json.Marshal(d)
}

func (d *BankCardData) GetMetaInfo() string {

	return "Банковская карта"
}

// Методы для BinaryData
func (d *BinaryData) ToBinary() ([]byte, error) {
	return d.Data, nil
}

func (d *BinaryData) GetMetaInfo() string {
	if d.Description != "" {
		return d.Description
	}
	return d.FileName
}

// Функции для createArtifactModel
func newCreateArtifactModel() createArtifactModel {
	return createArtifactModel{
		screen: newCreateArtifactScreen(),
	}
}

func (m createArtifactModel) Init() tea.Cmd {
	return m.screen.Init()
}

func (m createArtifactModel) Update(msg tea.Msg) (createArtifactModel, tea.Cmd) {
	var cmd tea.Cmd
	m.screen, cmd = m.screen.Update(msg)
	return m, cmd
}

func (m createArtifactModel) View() string {
	return m.screen.View()
}

// Функции для createArtifactScreen
func newCreateArtifactScreen() createArtifactScreen {
	// Список типов артефактов
	items := []list.Item{
		artifactTypeItem{name: "📝 Текст", desc: "Произвольный текст", typ: proto.ArtifactTypeEnum_TEXT},
		artifactTypeItem{name: "🔐 Логин/Пароль", desc: "Учетные данные", typ: proto.ArtifactTypeEnum_LOGIN_PASSWORD},
		artifactTypeItem{name: "💳 Банковская карта", desc: "Данные банковской карты", typ: proto.ArtifactTypeEnum_BANK_CARD},
		artifactTypeItem{name: "📎 Файл", desc: "Бинарный файл", typ: proto.ArtifactTypeEnum_BINARY},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("170")).
		BorderLeftForeground(lipgloss.Color("170"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("170")).
		BorderLeftForeground(lipgloss.Color("170"))

	typeList := list.New(items, delegate, 0, 0)
	typeList.Title = "Выберите тип артефакта"
	typeList.Styles.Title = titleStyle

	return createArtifactScreen{
		state:        selectTypeScreenState,
		typeSelector: typeList,
		inputs:       make(map[string]textinput.Model),
	}
}

func (s createArtifactScreen) Init() tea.Cmd {
	return nil
}

func (s createArtifactScreen) Update(msg tea.Msg) (createArtifactScreen, tea.Cmd) {
	switch s.state {
	case selectTypeScreenState:
		return s.updateTypeSelection(msg)
	case fillDataScreenState:
		return s.updateDataInput(msg)
	}
	return s, nil
}

func (s createArtifactScreen) updateTypeSelection(msg tea.Msg) (createArtifactScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return s, func() tea.Msg { return navigateToMsg{state: mainMenuState} }
		case "enter":
			if selected, ok := s.typeSelector.SelectedItem().(artifactTypeItem); ok {
				s.artifactType = selected.typ
				s.state = fillDataScreenState
				s.initializeInputs()
			}
		}
	}

	var cmd tea.Cmd
	s.typeSelector, cmd = s.typeSelector.Update(msg)
	return s, cmd
}

func (s createArtifactScreen) updateDataInput(msg tea.Msg) (createArtifactScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			s.state = selectTypeScreenState
			return s, nil
		case "tab", "shift+tab", "up", "down":
			// Навигация по полям
			key := msg.String()
			fieldNames := s.getFieldNames()

			if key == "up" || key == "shift+tab" {
				s.focusIndex--
			} else {
				s.focusIndex++
			}

			if s.focusIndex >= len(fieldNames) {
				s.focusIndex = 0
			} else if s.focusIndex < 0 {
				s.focusIndex = len(fieldNames) - 1
			}

			for i, fieldName := range fieldNames {
				model := s.inputs[fieldName]
				if i == s.focusIndex {
					model.Focus()
				} else {
					model.Blur()
				}
				s.inputs[fieldName] = model
			}
			return s, nil
		case "enter":
			if s.submitting {
				return s, nil
			}
			// Если нажат Enter и валидация пройдена - создаем артефакт
			if s.validateInputs() {
				s.submitting = true
				return s, s.submitArtifact()
			} else {
				// Если валидация не пройдена, показываем сообщение об ошибке
				logger.Info("Валидация не пройдена, обязательные поля не заполнены")
			}
		}
	}

	// Обновляем активное поле ввода
	fieldNames := s.getFieldNames()
	if s.focusIndex < len(fieldNames) {
		activeField := fieldNames[s.focusIndex]
		var cmd tea.Cmd
		s.inputs[activeField], cmd = s.inputs[activeField].Update(msg)
		return s, cmd
	}

	return s, nil
}

func (s *createArtifactScreen) initializeInputs() {
	s.inputs = make(map[string]textinput.Model)
	s.focusIndex = 0 // Сбрасываем индекс фокуса

	switch s.artifactType {
	case proto.ArtifactTypeEnum_TEXT:
		s.inputs["title"] = createInput("Название", "")
		s.inputs["content"] = createInput("Содержимое", "")

	case proto.ArtifactTypeEnum_LOGIN_PASSWORD:
		s.inputs["site"] = createInput("Сайт", "")
		s.inputs["login"] = createInput("Логин", "")
		s.inputs["password"] = createInput("Пароль", "")
		s.inputs["notes"] = createInput("Заметки (опционально)", "")

	case proto.ArtifactTypeEnum_BANK_CARD:
		s.inputs["number"] = createInput("Номер карты", "")
		s.inputs["holder"] = createInput("Держатель карты", "")
		s.inputs["expiry"] = createInput("Срок действия (MM/YY)", "")
		s.inputs["cvv"] = createInput("CVV (опционально)", "")

	case proto.ArtifactTypeEnum_BINARY:
		s.inputs["filepath"] = createInput("Путь к файлу", "")
		s.inputs["description"] = createInput("Описание (опционально)", "")
	}

	if len(s.inputs) > 0 {
		fieldNames := s.getFieldNames()
		if len(fieldNames) > 0 {
			// Сначала убираем фокус у всех полей
			for _, fieldName := range fieldNames {
				model := s.inputs[fieldName]
				model.Blur()
				s.inputs[fieldName] = model
			}
			// Затем фокусируем первое поле
			model := s.inputs[fieldNames[0]]
			model.Focus()
			s.inputs[fieldNames[0]] = model
		}
	}
}

func createInput(placeholder, value string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(value)
	ti.Width = 40
	return ti
}

func (s createArtifactScreen) getFieldNames() []string {
	var names []string
	switch s.artifactType {
	case proto.ArtifactTypeEnum_TEXT:
		names = []string{"title", "content"}
	case proto.ArtifactTypeEnum_LOGIN_PASSWORD:
		names = []string{"site", "login", "password", "notes"}
	case proto.ArtifactTypeEnum_BANK_CARD:
		names = []string{"number", "holder", "expiry", "cvv"}
	case proto.ArtifactTypeEnum_BINARY:
		names = []string{"filepath", "description"}
	}
	return names
}

func (s createArtifactScreen) validateInputs() bool {
	// Базовая валидация обязательных полей
	switch s.artifactType {
	case proto.ArtifactTypeEnum_TEXT:
		return strings.TrimSpace(s.inputs["content"].Value()) != ""
	case proto.ArtifactTypeEnum_LOGIN_PASSWORD:
		return strings.TrimSpace(s.inputs["login"].Value()) != "" &&
			strings.TrimSpace(s.inputs["password"].Value()) != ""
	case proto.ArtifactTypeEnum_BANK_CARD:
		return strings.TrimSpace(s.inputs["number"].Value()) != "" &&
			strings.TrimSpace(s.inputs["holder"].Value()) != "" &&
			strings.TrimSpace(s.inputs["expiry"].Value()) != ""
	case proto.ArtifactTypeEnum_BINARY:
		filePath := strings.TrimSpace(s.inputs["filepath"].Value())
		if filePath == "" {
			return false
		}
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func (s createArtifactScreen) submitArtifact() tea.Cmd {
	logger.Info("Начало создания артефакта", "type", s.artifactType)

	return func() tea.Msg {
		var payload []byte
		var metaInfo string
		var err error

		switch s.artifactType {
		case proto.ArtifactTypeEnum_TEXT:
			data := &TextData{
				Title:   s.inputs["title"].Value(),
				Content: s.inputs["content"].Value(),
			}
			payload, err = data.ToBinary()
			metaInfo = data.GetMetaInfo()

		case proto.ArtifactTypeEnum_LOGIN_PASSWORD:
			data := &LoginPasswordData{
				Site:     s.inputs["site"].Value(),
				Login:    s.inputs["login"].Value(),
				Password: s.inputs["password"].Value(),
				Notes:    s.inputs["notes"].Value(),
			}
			payload, err = data.ToBinary()
			metaInfo = data.GetMetaInfo()

		case proto.ArtifactTypeEnum_BANK_CARD:
			n, err := strconv.ParseUint(s.inputs["number"].Value(), 10, 64)
			if err != nil {
				logger.Error("Ошибка парсинга номера карты:", err)
			}
			cvv, err := strconv.ParseUint(s.inputs["cvv"].Value(), 10, 64)
			if err != nil {
				logger.Error("Ошибка парсинга CVV:", err)
			}
			data := &BankCardData{
				Number:     n,
				Holder:     s.inputs["holder"].Value(),
				ExpiryDate: s.inputs["expiry"].Value(),
				CVV:        cvv,
			}
			payload, err = data.ToBinary()
			metaInfo = data.GetMetaInfo()

		case proto.ArtifactTypeEnum_BINARY:
			filePath := s.inputs["filepath"].Value()
			fileData, err := os.ReadFile(filePath)
			if err != nil {
				return artifactCreateResultMsg{
					success: false,
					error:   fmt.Errorf("ошибка чтения файла: %v", err),
				}
			}
			data := &BinaryData{
				FileName:    filePath,
				Description: s.inputs["description"].Value(),
				Data:        fileData,
			}
			payload, err = data.ToBinary()
			if err != nil {
				logger.Error("Ошибка подготовки данных для бинарного файла:", err)
			}
			metaInfo = data.GetMetaInfo()
		}
		if err != nil {
			logger.Error("Ошибка подготовки данных:", err)
		}
		logger.Info("Артефакт подготовлен", "metaInfo", metaInfo)

		if err != nil {
			return artifactCreateResultMsg{
				success: false,
				error:   fmt.Errorf("ошибка подготовки данных: %v", err),
			}
		}

		req := &proto.CreateArtifactRequest{
			Type:     s.artifactType,
			MetaInfo: metaInfo,
			Payload:  payload,
		}

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

func (s createArtifactScreen) View() string {
	switch s.state {
	case selectTypeScreenState:
		view := s.typeSelector.View()
		// Добавляем информацию о выбранном элементе
		if selected, ok := s.typeSelector.SelectedItem().(artifactTypeItem); ok {
			view += fmt.Sprintf("\n\n📋 Выбран: %s - %s", selected.name, selected.desc)
		}
		return view + "\n\n↑/↓ - навигация • Enter - выбор • Esc - назад"

	case fillDataScreenState:
		var b strings.Builder
		b.WriteString("➕ Создание артефакта\n\n")

		// Показываем тип создаваемого артефакта
		typeName := ""
		switch s.artifactType {
		case proto.ArtifactTypeEnum_TEXT:
			typeName = "📝 Текст"
		case proto.ArtifactTypeEnum_LOGIN_PASSWORD:
			typeName = "🔐 Логин/Пароль"
		case proto.ArtifactTypeEnum_BANK_CARD:
			typeName = "💳 Банковская карта"
		case proto.ArtifactTypeEnum_BINARY:
			typeName = "📎 Файл"
		}
		b.WriteString(fmt.Sprintf("Тип: %s\n", typeName))
		b.WriteString(fmt.Sprintf("Активное поле: %d\n\n", s.focusIndex)) // Отладочная информация

		if s.submitting {
			b.WriteString("⏳ Создание...\n\n")
			return b.String()
		}

		fieldNames := s.getFieldNames()
		for i, fieldName := range fieldNames {
			input := s.inputs[fieldName]
			// Показываем курсор для активного поля
			cursor := " "
			if i == s.focusIndex {
				cursor = "▶"
			}

			// Показываем, какое поле активно
			status := " "
			if input.Focused() {
				status = "●"
			}

			b.WriteString(fmt.Sprintf("%s [%s] %s:\n%s\n\n", cursor, status, input.Placeholder, input.View()))
		}

		b.WriteString("Enter - создать • Esc - назад • Tab - переключение полей\n")

		if !s.validateInputs() {
			b.WriteString("\n⚠️  Заполните обязательные поля")
		}

		return b.String()
	}

	return ""
}
