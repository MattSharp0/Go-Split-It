package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// LogLevel represents the logging level
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// LogOutput represents where logs should be written
type LogOutput string

const (
	OutputStdout LogOutput = "stdout"
	OutputFile   LogOutput = "file"
	OutputBoth   LogOutput = "both"
)

// Config holds logger configuration
type Config struct {
	Level      LogLevel
	Output     LogOutput
	FilePath   string
	JSONFormat bool
}

var (
	defaultLogger *slog.Logger
)

// customHandler wraps slog.Handler to provide custom formatting
type customHandler struct {
	opts slog.HandlerOptions
	w    io.Writer
	mu   *sync.Mutex
}

func newCustomHandler(w io.Writer, opts *slog.HandlerOptions) *customHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &customHandler{
		opts: *opts,
		w:    w,
		mu:   &sync.Mutex{},
	}
}

func (h *customHandler) Enabled(ctx context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

func (h *customHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	buf := make([]byte, 0, 1024)

	// Format: 2025-10-20 14:30:15.123
	t := r.Time.Format("2006-01-02 15:04:05.000")
	buf = append(buf, t...)
	buf = append(buf, ' ')

	// Level with padding for alignment
	level := r.Level.String()
	// Pad level to 5 characters for alignment (DEBUG/INFO /WARN /ERROR)
	if len(level) < 5 {
		level = level + strings.Repeat(" ", 5-len(level))
	}
	buf = append(buf, level...)
	buf = append(buf, ' ')

	// Source information (file:line) if available
	if h.opts.AddSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()

		// Extract just the filename (not full path)
		file := filepath.Base(f.File)

		buf = append(buf, '[')
		buf = append(buf, file...)
		buf = append(buf, ':')
		buf = append(buf, fmt.Sprintf("%d", f.Line)...)
		buf = append(buf, "] "...)
	}

	// Message
	buf = append(buf, r.Message...)

	// Attributes
	if r.NumAttrs() > 0 {
		r.Attrs(func(a slog.Attr) bool {
			buf = append(buf, ' ')
			buf = append(buf, a.Key...)
			buf = append(buf, '=')
			buf = append(buf, a.Value.String()...)
			return true
		})
	}

	buf = append(buf, '\n')

	_, err := h.w.Write(buf)
	return err
}

func (h *customHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *customHandler) WithGroup(name string) slog.Handler {
	return h
}

// InitLogger initializes the global logger with the given configuration
func InitLogger(cfg Config) (*slog.Logger, error) {
	// Determine log level
	var level slog.Level
	switch cfg.Level {
	case LevelDebug:
		level = slog.LevelDebug
	case LevelInfo:
		level = slog.LevelInfo
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Determine output destination
	var writer io.Writer
	var err error

	switch cfg.Output {
	case OutputStdout:
		writer = os.Stdout
	case OutputFile:
		writer, err = openLogFile(cfg.FilePath)
		if err != nil {
			return nil, err
		}
	case OutputBoth:
		fileWriter, err := openLogFile(cfg.FilePath)
		if err != nil {
			return nil, err
		}
		writer = io.MultiWriter(os.Stdout, fileWriter)
	default:
		writer = os.Stdout
	}

	// Create handler based on format preference
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true, // Add source file/line for all levels
	}

	if cfg.JSONFormat {
		handler = slog.NewJSONHandler(writer, opts)
	} else {
		handler = newCustomHandler(writer, opts) // Use custom handler
	}

	logger := slog.New(handler)

	// Set as default logger
	defaultLogger = logger
	slog.SetDefault(logger)

	return logger, nil
}

// openLogFile opens or creates a log file
func openLogFile(filePath string) (*os.File, error) {
	if filePath == "" {
		filePath = "logs/app.log"
	}

	// Create logs directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Open file in append mode, create if doesn't exist
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// LoadConfigFromEnv loads logger configuration from environment variables
func LoadConfigFromEnv() Config {
	cfg := Config{
		Level:      LevelInfo, // default
		Output:     OutputStdout,
		FilePath:   os.Getenv("LOG_FILE_PATH"),
		JSONFormat: false,
	}

	// Parse log level
	if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" {
		cfg.Level = LogLevel(strings.ToLower(levelStr))
	}

	// Parse log output
	if outputStr := os.Getenv("LOG_OUTPUT"); outputStr != "" {
		cfg.Output = LogOutput(strings.ToLower(outputStr))
	}

	// Parse log format
	if formatStr := os.Getenv("LOG_FORMAT"); formatStr != "" {
		cfg.JSONFormat = strings.ToLower(formatStr) == "json"
	}

	return cfg
}

// Get returns the default logger instance
func Get() *slog.Logger {
	if defaultLogger == nil {
		// Initialize with defaults if not already initialized
		defaultLogger = slog.Default()
	}
	return defaultLogger
}

// Helper functions for common log operations

// Debug logs a debug message
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// With returns a logger with the given attributes
func With(args ...any) *slog.Logger {
	return Get().With(args...)
}
