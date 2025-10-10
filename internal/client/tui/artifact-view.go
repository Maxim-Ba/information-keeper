package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Maxim-Ba/information-keeper/pkg/logger"
	"github.com/Maxim-Ba/information-keeper/pkg/proto"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewArtifactDetailsModel struct {
	artifact *proto.Artifact
	content  string
	loading  bool
	width    int
	height   int
	error    string
}

func newViewArtifactDetailsModel(artifact *proto.Artifact) viewArtifactDetailsModel {
	return viewArtifactDetailsModel{
		artifact: artifact,
		loading:  true,
	}
}

func (m *viewArtifactDetailsModel) Init() tea.Cmd {
	if m.artifact == nil {
		logger.Error("❌ ОШИБКА: artifact is nil в viewArtifactDetailsModel.Init()")
		m.loading = false
		m.error = "Ошибка: артефакт не найден"
		return nil // Возвращаем nil вместо отправки сообщения
	}

	logger.Info("🚀 Инициализация viewArtifactDetailsModel",
		"artifactID", m.artifact.Id,
		"loading", m.loading)
	return m.loadArtifactContent
}

func (m *viewArtifactDetailsModel) Update(msg tea.Msg) (viewArtifactDetailsModel, tea.Cmd) {
	logger.Info("🔄 viewArtifactDetailsModel Update", "type", fmt.Sprintf("%T", msg))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		logger.Info("⌨️ Обработка клавиши", "key", msg.String())
		switch msg.String() {
		case ESC, "q":
			return *m, func() tea.Msg { return navigateToMsg{state: artifactDetailState} }
		case "c":
			// Копирование ссылки для бинарных файлов
			if m.artifact.Type == proto.ArtifactTypeEnum_BINARY && m.artifact.Link != "" {
				if err := copyToClipboard(m.artifact.Link); err != nil {
					logger.Error("❌ Ошибка копирования в буфер", "error", err)
					return *m, func() tea.Msg {
						displayURL := m.artifact.Link
						if len(displayURL) > 100 {
							displayURL = displayURL[:97] + "..."
						}
						return successMsg(fmt.Sprintf("❌ Не удалось скопировать автоматически.\nСкопируйте ссылку вручную:\n%s", displayURL))
					}
				}
				return *m, func() tea.Msg {
					return successMsg("✅ Ссылка скопирована в буфер обмена: " + m.artifact.Link)
				}
			}
		}
	case artifactContentLoadedMsg:
		logger.Info("✅ Содержимое артефакта загружено")
		m.loading = false
		m.content = msg.content
	case artifactContentErrorMsg:
		logger.Info("❌ Ошибка загрузки содержимого", "error", msg.error)
		m.loading = false
		m.error = msg.error
	}
	return *m, nil
}

func (m *viewArtifactDetailsModel) View() string {
	if m.artifact == nil {
		return errorStyle.Render("❌ Ошибка: артефакт не найден")
	}
	logger.Info("🖥️ Отображение деталей артефакта",
		"loading", m.loading,
		"hasError", m.error != "",
		"hasContent", m.content != "",
		"contentLength", len(m.content),
		"artifactType", m.artifact.Type)
	var b strings.Builder

	b.WriteString("📋 Детали артефакта\n\n")
	b.WriteString(fmt.Sprintf("Название: %s\n", m.artifact.MetaInfo))
	b.WriteString(fmt.Sprintf("Тип: %s\n", getArtifactTypeName(m.artifact.Type)))
	b.WriteString(fmt.Sprintf("Создан: %s\n", time.Unix(m.artifact.CreatedAt, 0).Format("02.01.2006 15:04")))
	b.WriteString(fmt.Sprintf("Обновлен: %s\n", time.Unix(m.artifact.UpdatedAt, 0).Format("02.01.2006 15:04")))

	if m.artifact.ExpiredAt > 0 {
		b.WriteString(fmt.Sprintf("Истекает: %s\n", time.Unix(m.artifact.ExpiredAt, 0).Format("02.01.2006 15:04")))
	}

	b.WriteString("\n")

	if m.loading {
		b.WriteString("⏳ Загрузка содержимого...\n")
	} else if m.error != "" {
		b.WriteString(fmt.Sprintf("❌ Ошибка: %s\n", m.error))
	} else {
		if m.artifact.Type == proto.ArtifactTypeEnum_BINARY {
			b.WriteString("📎 Бинарный файл\n")
			if m.artifact.Link != "" && strings.HasPrefix(m.artifact.Link, "http") {
				// Всегда показываем полную ссылку
				b.WriteString(fmt.Sprintf("Ссылка для скачивания:\n%s\n", m.artifact.Link))
				b.WriteString("\nНажмите 'c' чтобы скопировать ссылку\n")
				b.WriteString("Или выделите и скопируйте ссылку выше вручную\n")
			} else {
				b.WriteString("⏳ Получение ссылки для скачивания...\n")
			}
		} else {
			// отображение содержимого для НЕ-бинарных типов
			b.WriteString("📄 Содержимое:\n")
			contentWidth := m.width - 6 // оставляем место для рамки
			if contentWidth < 10 {
				contentWidth = 10
			}
			formattedContent := formatContent(m.content, contentWidth)
			b.WriteString(formattedContent)
		}
	}

	b.WriteString("\n\nEsc - назад")
	if m.artifact.Type == proto.ArtifactTypeEnum_BINARY && m.artifact.Link != "" {
		b.WriteString(" • c - копировать ссылку")
	}

	content := b.String()
	style := lipgloss.NewStyle().
		Width(m.width - 4).
		MaxWidth(80).
		Height(m.height - 4).
		Align(lipgloss.Left)

	return style.Render(content)
}
func (m *viewArtifactDetailsModel) loadArtifactContent() tea.Msg {
	logger.Info("🔄 Начало загрузки содержимого артефакта",
		"artifactID", m.artifact.Id,
		"type", m.artifact.Type,
		"link", m.artifact.Link)

	if m.artifact.Type == proto.ArtifactTypeEnum_BINARY {
		logger.Info("📎 Бинарный файл - получаем presigned URL")

		// Для бинарных файлов получаем presigned URL
		downloadURL, err := globalClient.GetArtifactDownloadURL(context.Background(), m.artifact.Id)
		if err != nil {
			logger.Error("❌ Ошибка получения download URL", "error", err)
			return artifactContentErrorMsg{error: fmt.Sprintf("❌ Ошибка получения ссылки: %v", err)}
		}
		logger.Info("downloadURL", downloadURL)
		// Сохраняем ссылку для отображения
		m.artifact.Link = downloadURL
		return artifactContentLoadedMsg{content: "binary"}
	}

	// Для остальных типов скачиваем содержимое через presigned URL
	downloadURL, err := globalClient.GetArtifactDownloadURL(context.Background(), m.artifact.Id)
	if err != nil {
		logger.Error("❌ Ошибка получения download URL", "error", err)
		return artifactContentErrorMsg{error: fmt.Sprintf("❌ Ошибка получения ссылки: %v", err)}
	}

	logger.Info("📥 Скачиваем содержимое по presigned URL", "url", downloadURL)
	ctx:= context.Background()
	content, err := globalClient.DownloadContent(ctx, downloadURL)
	if err != nil {
		logger.Error("❌ Ошибка загрузки содержимого", "error", err)
		return artifactContentErrorMsg{error: fmt.Sprintf("❌ Ошибка загрузки: %v", err)}
	}

	parsedContent, err := parseDownloadedContent(m.artifact.Type, content)
	if err != nil {
		logger.Error("❌ Ошибка парсинга содержимого", "error", err)
		return artifactContentErrorMsg{error: fmt.Sprintf("❌ Ошибка парсинга: %v", err)}
	}

	logger.Info("✅ Содержимое успешно распарсено")
	return artifactContentLoadedMsg{content: parsedContent}
}

func parseDownloadedContent(artifactType proto.ArtifactTypeEnum, content []byte) (string, error) {
	switch artifactType {
	case proto.ArtifactTypeEnum_TEXT:
		var data TextData
		if err := json.Unmarshal(content, &data); err != nil {
			// Если не удалось распарсить как JSON, показываем как plain text
			return string(content), err
		}
		if data.Title != "" {
			return fmt.Sprintf("Заголовок: %s\n\n%s", data.Title, data.Content), nil
		}
		return data.Content, nil

	case proto.ArtifactTypeEnum_LOGIN_PASSWORD:
		var data LoginPasswordData
		if err := json.Unmarshal(content, &data); err != nil {
			return string(content), nil
		}
		result := fmt.Sprintf("Логин: %s\nПароль: %s", data.Login, data.Password)
		if data.Site != "" {
			result = fmt.Sprintf("Сайт: %s\n%s", data.Site, result)
		}
		if data.Notes != "" {
			result = fmt.Sprintf("%s\nЗаметки: %s", result, data.Notes)
		}
		return result, nil

	case proto.ArtifactTypeEnum_BANK_CARD:
		var data BankCardData
		if err := json.Unmarshal(content, &data); err != nil {
			return string(content), nil
		}
		result := fmt.Sprintf("Номер карты: %d\nДержатель: %s\nСрок действия: %s",
			data.Number, data.Holder, data.ExpiryDate)
		if data.CVV != 0 {
			result = fmt.Sprintf("%s\nCVV: %d", result, data.CVV)
		}
		return result, nil

	default:
		return string(content), nil
	}
}

func formatContent(content string, maxWidth int) string {
	lines := strings.Split(content, "\n")
	var formatted strings.Builder

	for _, line := range lines {
		// Перенос длинных строк
		for len(line) > maxWidth {
			formatted.WriteString("│ " + line[:maxWidth] + " │\n")
			line = line[maxWidth:]
		}
		if line != "" {
			formatted.WriteString("│ " + line + strings.Repeat(" ", maxWidth-len(line)) + " │\n")
		}
	}

	return formatted.String()
}

type artifactContentLoadedMsg struct {
	content string
}

type artifactContentErrorMsg struct {
	error string
}
