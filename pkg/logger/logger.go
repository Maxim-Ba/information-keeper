// logger методы определения Logger
package logger

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel представляет уровни логирования.
type LogLevel slog.Level

const (
	// LevelDebug представляет уровни логирования.
	LevelDebug LogLevel = LogLevel(slog.LevelDebug)
	// LevelInfo  LogLevel = LogLevel(slog.LevelInfo)
	LevelInfo  LogLevel = LogLevel(slog.LevelInfo)
	// LevelWarn  LogLevel = LogLevel(slog.LevelWarn)
	LevelWarn  LogLevel = LogLevel(slog.LevelWarn)
	// LevelError LogLevel = LogLevel(slog.LevelError)
	LevelError LogLevel = LogLevel(slog.LevelError)
)

// Config конфигурация логгера.
type Config struct {
	FilePath    string
	Level       LogLevel
	UseJSON     bool
	AddSource   bool
	MaxFileSize int64 // в байтах
}

// FileLogger структура логгера с ротацией.
type FileLogger struct {
	handler slog.Handler
	file    *os.File
	mu      sync.Mutex
	config  Config
}

var (
	globalLogger *slog.Logger
	fileLogger   *FileLogger
)
// DefaultConfig возвращает конфигурацию логгера по умолчанию
func DefaultConfig() Config {
	return Config{
		FilePath:    "",
		Level:       LevelInfo,
		UseJSON:     false,
		AddSource:   true,
		MaxFileSize: 10 * 1024 * 1024, // 10MB
	}
}
// InitLogger инициализирует глобальный логгер с заданной конфигурацией
func InitLogger(config Config) error {
	fl, err := NewFileLogger(config)
	if err != nil {
		return err
	}

	fileLogger = fl
	globalLogger = slog.New(fl.handler)

	slog.SetDefault(globalLogger)

	return nil
}
// NewFileLogger создает новый файловый логгер с поддержкой ротации
func NewFileLogger(config Config) (*FileLogger, error) {
	// Если FilePath не задан, используем stdout
	if config.FilePath == "" {
		opts := &slog.HandlerOptions{
			Level:     slog.Level(config.Level),
			AddSource: config.AddSource,
		}

		var handler slog.Handler
		if config.UseJSON {
			handler = slog.NewJSONHandler(os.Stdout, opts)
		} else {
			handler = slog.NewTextHandler(os.Stdout, opts)
		}

		return &FileLogger{
			handler: handler,
			file:    nil, // stdout не требует закрытия
			config:  config,
		}, nil
	}

	// Создаем директорию для файла лога, если она не существует
	if err := os.MkdirAll(filepath.Dir(config.FilePath), 0750); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}

	opts := &slog.HandlerOptions{
		Level:     slog.Level(config.Level),
		AddSource: config.AddSource,
	}

	var handler slog.Handler
	if config.UseJSON {
		handler = slog.NewJSONHandler(file, opts)
	} else {
		handler = slog.NewTextHandler(file, opts)
	}

	return &FileLogger{
		handler: handler,
		file:    file,
		config:  config,
	}, nil
}
// Close закрывает файловый логгер и освобождает ресурсы
func (fl *FileLogger) Close() error {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	// Если файл не nil (не stdout), закрываем его
	if fl.file != nil {
		return fl.file.Close()
	}
	return nil
}

func (fl *FileLogger) shouldRotate() bool {
	// Для stdout ротация не нужна
	if fl.file == nil || fl.config.MaxFileSize <= 0 {
		return false
	}

	info, err := fl.file.Stat()
	if err != nil {
		return false
	}

	return info.Size() >= fl.config.MaxFileSize
}

func (fl *FileLogger) rotate() error {
	// Для stdout ротация не нужна
	if fl.file == nil {
		return nil
	}

	fl.mu.Lock()
	defer fl.mu.Unlock()

	if fl.file == nil {
		return nil
	}

	if err := fl.file.Close(); err != nil {
		return err
	}

	timestamp := time.Now().Format("20060102_150405")
	backupFile := fl.config.FilePath + "." + timestamp
	if err := os.Rename(fl.config.FilePath, backupFile); err != nil {
		return err
	}

	file, err := os.OpenFile(fl.config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}

	opts := &slog.HandlerOptions{
		Level:     slog.Level(fl.config.Level),
		AddSource: fl.config.AddSource,
	}

	if fl.config.UseJSON {
		fl.handler = slog.NewJSONHandler(file, opts)
	} else {
		fl.handler = slog.NewTextHandler(file, opts)
	}

	fl.file = file
	return nil
}
// Handle обрабатывает запись лога
func (fl *FileLogger) Handle(ctx context.Context, r slog.Record) error {
	if fl.shouldRotate() {
		if err := fl.rotate(); err != nil {
			slog.Error("Failed to rotate log file", "error", err)
		}
	}

	return fl.handler.Handle(ctx, r)
}
// WithAttrs создает новый логгер с дополнительными атрибутами
func (fl *FileLogger) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &FileLogger{
		handler: fl.handler.WithAttrs(attrs),
		file:    fl.file,
		config:  fl.config,
	}
}

// WithGroup создает новый логгер с указанной группой
func (fl *FileLogger) WithGroup(name string) slog.Handler {
	return &FileLogger{
		handler: fl.handler.WithGroup(name),
		file:    fl.file,
		config:  fl.config,
	}
}
// Enabled проверяет, включен ли указанный уровень логирования
func (fl *FileLogger) Enabled(ctx context.Context, level slog.Level) bool {
	return fl.handler.Enabled(ctx, level)
}
// GetLogger возвращает глобальный экземпляр логгера
func GetLogger() *slog.Logger {
	return globalLogger
}
// CloseLogger закрывает файловый логгер
func CloseLogger() {
	if fileLogger != nil {
		fileLogger.Close()
	}
}

// Debug записывает сообщение отладочного уровня
func Debug(msg string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(msg, args...)
	}
}

// Info записывает информационное сообщение
func Info(msg string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Info(msg, args...)
	}
}
// Warn записывает предупреждающее сообщение
func Warn(msg string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(msg, args...)
	}
}

// Error записывает сообщение об ошибке
func Error(msg string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Error(msg, args...)
	}
}

// ErrorContext записывает сообщение об ошибке с контекстом
func ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.ErrorContext(ctx, msg, args...)
	}
}
