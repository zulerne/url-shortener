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

	"github.com/zulerne/url-shortener/internal/config"
)

const (
	reset  = "\033[0m"
	gray   = "\033[90m"
	blue   = "\033[34m"
	yellow = "\033[33m"
	red    = "\033[31m"
)

// Handler is a custom slog.Handler with colored output and stack traces.
type Handler struct {
	out     io.Writer
	wd      string
	attrs   []slog.Attr
	groups  []string
	level   slog.Level
	noColor bool
}

// bufferPool is used to avoid allocations in the hot path.
var bufferPool = sync.Pool{
	New: func() any {
		return new(strings.Builder)
	},
}

// Option configures a LoggerHandler.
type Option func(*Handler)

// WithLevel sets the minimum log level.
func WithLevel(level slog.Level) Option {
	return func(h *Handler) {
		h.level = level
	}
}

// WithNoColor disables colored output (useful for production/files).
func WithNoColor(noColor bool) Option {
	return func(h *Handler) {
		h.noColor = noColor
	}
}

// NewLoggerHandler creates a new LoggerHandler with the given options.
func NewLoggerHandler(out io.Writer, opts ...Option) *Handler {
	wd, _ := os.Getwd()
	h := &Handler{
		out:   out,
		wd:    wd,
		level: slog.LevelDebug, // Default: log everything
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func SetupLogger(env string) {
	var log *slog.Logger

	switch env {
	case config.EnvLocal:
		log = slog.New(NewLoggerHandler(os.Stdout, WithNoColor(false), WithLevel(slog.LevelDebug)))
	case config.EnvDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case config.EnvProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	slog.SetDefault(log)
}

// Enabled reports whether the handler handles records at the given level.
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// WithAttrs returns a new handler with the given attributes added.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Create a new slice to avoid sharing the underlying array
	newAttrs := make([]slog.Attr, len(h.attrs), len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	newAttrs = append(newAttrs, attrs...)

	return &Handler{
		out:     h.out,
		wd:      h.wd,
		attrs:   newAttrs,
		groups:  h.groups,
		level:   h.level,
		noColor: h.noColor,
	}
}

// WithGroup returns a new handler with the given group name added.
func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	newGroups := make([]string, len(h.groups), len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups = append(newGroups, name)

	return &Handler{
		out:     h.out,
		wd:      h.wd,
		attrs:   h.attrs,
		groups:  newGroups,
		level:   h.level,
		noColor: h.noColor,
	}
}

// Handle handles the log record.
func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	// Get buffer from pool to avoid allocations
	buf := bufferPool.Get().(*strings.Builder)
	buf.Reset()
	defer bufferPool.Put(buf)

	// Use fixed-width format (6 digits for microseconds) to prevent alignment issues
	timeStr := r.Time.Format("2006-01-02T15:04:05.000000-07:00")

	levelColor := h.color(gray)
	switch r.Level {
	case slog.LevelDebug:
		levelColor = h.color(gray)
	case slog.LevelInfo:
		levelColor = h.color(blue)
	case slog.LevelWarn:
		levelColor = h.color(yellow)
	case slog.LevelError:
		levelColor = h.color(red)
	}

	fs := runtime.CallersFrames([]uintptr{r.PC})
	frame, _ := fs.Next()
	sourcePath := frame.File
	if h.wd != "" {
		if rel, err := filepath.Rel(h.wd, sourcePath); err == nil {
			sourcePath = rel
		}
	}
	source := fmt.Sprintf("%s:%d", sourcePath, frame.Line)

	// Add prefix for groups
	prefix := ""
	if len(h.groups) > 0 {
		prefix = strings.Join(h.groups, ".") + "."
	}

	// Build complete log line in buffer
	fmt.Fprintf(buf, "%s\t%s%-7s%s\t%s%-20s%s\t%s",
		timeStr, levelColor, r.Level.String(), h.color(reset),
		h.color(gray), source, h.color(reset),
		r.Message)

	// Write attrs directly to buf (avoids extra allocation)
	for _, a := range h.attrs {
		fmt.Fprintf(buf, " %s%s%s=%v%s", h.color(gray), prefix, a.Key, a.Value, h.color(reset))
	}
	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(buf, " %s%s%s=%v%s", h.color(gray), prefix, a.Key, a.Value, h.color(reset))
		return true
	})
	buf.WriteByte('\n')

	// Single write to output - io.Writer is expected to be thread-safe
	// (os.Stdout, os.Stderr, and most file writers are thread-safe)
	_, err := io.WriteString(h.out, buf.String())
	return err
}

// color returns the color code if colors are enabled, empty string otherwise.
func (h *Handler) color(c string) string {
	if h.noColor {
		return ""
	}
	return c
}
