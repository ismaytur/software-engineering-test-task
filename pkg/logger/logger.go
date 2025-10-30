package logger

import (
	"context"
	"errors"
	"fmt"
	"io"
	stdlog "log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	OutputStdout = "stdout"
	OutputFile   = "file"
	OutputBoth   = "both"
)

type Options struct {
	Output   string
	FilePath string
	Level    string
}

type Logger struct {
	base    *slog.Logger
	closers []io.Closer
}

type ctxKey struct{}

var (
	global     atomic.Pointer[Logger]
	configLock sync.Mutex
)

func DefaultOptions() Options {
	return Options{
		Output: OutputStdout,
		Level:  "info",
	}
}

func Configure(opts Options) (*Logger, error) {
	configLock.Lock()
	defer configLock.Unlock()

	opts = normalizeOptions(opts)

	inst, err := newLogger(opts)
	if err != nil {
		return nil, err
	}

	// Replace global logger.
	prev := global.Swap(inst)
	if prev != nil {
		if err := prev.Close(); err != nil {
			inst.base.Error("failed to close previous logger", slog.String("error", err.Error()))
		}
	}

	// Bridge standard library log and slog defaults.
	stdlog.SetOutput(Writer(inst, slog.LevelInfo))
	stdlog.SetFlags(0)
	slog.SetDefault(inst.base)

	return inst, nil
}

func Get() *Logger {
	inst := global.Load()
	if inst != nil {
		return inst
	}
	l, err := Configure(DefaultOptions())
	if err != nil {
		panic(fmt.Sprintf("failed to configure default logger: %v", err))
	}
	return l
}

func (l *Logger) Base() *slog.Logger {
	return l.base
}

func (l *Logger) Info(msg string, attrs ...any) {
	l.base.Info(msg, attrs...)
}

func (l *Logger) Warn(msg string, attrs ...any) {
	l.base.Warn(msg, attrs...)
}

func (l *Logger) Error(msg string, attrs ...any) {
	l.base.Error(msg, attrs...)
}

func (l *Logger) Debug(msg string, attrs ...any) {
	l.base.Debug(msg, attrs...)
}

func (l *Logger) With(attrs ...any) *Logger {
	return &Logger{
		base:    l.base.With(attrs...),
		closers: l.closers,
	}
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	if ctxLogger, ok := ctx.Value(ctxKey{}).(*Logger); ok {
		return ctxLogger
	}
	return l
}

func ContextWithLogger(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

func (l *Logger) Close() error {
	var errs []error
	seen := map[io.Closer]struct{}{}
	for _, closer := range l.closers {
		if closer == nil {
			continue
		}
		if _, ok := seen[closer]; ok {
			continue
		}
		seen[closer] = struct{}{}
		if err := closer.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func Writer(l *Logger, level slog.Leveler) io.Writer {
	return &slogWriter{
		logger: l,
		level:  level,
	}
}

type slogWriter struct {
	logger *Logger
	level  slog.Leveler
}

func (w *slogWriter) Write(p []byte) (int, error) {
	msg := strings.TrimSpace(string(p))
	switch lvl := w.level.Level(); {
	case lvl <= slog.LevelDebug:
		w.logger.Debug(msg)
	case lvl <= slog.LevelInfo:
		w.logger.Info(msg)
	case lvl <= slog.LevelWarn:
		w.logger.Warn(msg)
	default:
		w.logger.Error(msg)
	}
	return len(p), nil
}

func newLogger(opts Options) (*Logger, error) {
	level, err := parseLevel(opts.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid log level %q, falling back to %q: %v\n", opts.Level, DefaultOptions().Level, err)
		level, _ = parseLevel(DefaultOptions().Level)
	}

	var writers []io.Writer
	var closers []io.Closer

	addFile := func(path string) error {
		if path == "" {
			return fmt.Errorf("file path cannot be empty when output includes file")
		}
		f, err := openLogFile(path)
		if err != nil {
			return fmt.Errorf("open log file: %w", err)
		}
		writers = append(writers, f)
		closers = append(closers, f)
		return nil
	}

	switch opts.Output {
	case OutputStdout:
		writers = append(writers, os.Stdout)
	case OutputFile:
		if err := addFile(opts.FilePath); err != nil {
			return nil, err
		}
	case OutputBoth:
		writers = append(writers, os.Stdout)
		if err := addFile(opts.FilePath); err != nil {
			return nil, err
		}
	default:
		writers = append(writers, os.Stdout)
	}

	handlerOpts := buildHandlerOptions(level)
	handler := slog.NewJSONHandler(io.MultiWriter(writers...), handlerOpts)

	return &Logger{
		base:    slog.New(handler),
		closers: closers,
	}, nil
}

func parseLevel(value string) (slog.Leveler, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return nil, fmt.Errorf("unknown log level: %s", value)
	}
}

func openLogFile(path string) (*os.File, error) {
	cleanPath := filepath.Clean(path)
	if !filepath.IsAbs(cleanPath) {
		return nil, fmt.Errorf("log file path must be absolute: %s", path)
	}
	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("ensure log directory: %w", err)
	}
	if info, err := os.Stat(cleanPath); err == nil {
		if info.IsDir() {
			return nil, fmt.Errorf("log file path points to a directory: %s", cleanPath)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("stat log file: %w", err)
	}
	return os.OpenFile(cleanPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
}

func buildHandlerOptions(level slog.Leveler) *slog.HandlerOptions {
	return &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			switch attr.Key {
			case slog.TimeKey:
				attr.Key = "timestamp"
			case slog.MessageKey:
				attr.Key = "message"
			}
			return attr
		},
	}
}

func normalizeOptions(opts Options) Options {
	defaults := DefaultOptions()

	opts.Output = cleanOption(opts.Output, defaults.Output)
	opts.Level = cleanOption(opts.Level, defaults.Level)
	opts.FilePath = strings.TrimSpace(opts.FilePath)

	return opts
}

func cleanOption(value, fallback string) string {
	value = strings.TrimSpace(value)
	if idx := strings.Index(value, "#"); idx >= 0 {
		value = strings.TrimSpace(value[:idx])
	}
	if value == "" {
		return fallback
	}
	return value
}

func OptionsFromEnv(env map[string]string) Options {
	return normalizeOptions(Options{
		Output:   env["LOG_OUTPUT"],
		FilePath: env["LOG_FILE"],
		Level:    env["LOG_LEVEL"],
	})
}
