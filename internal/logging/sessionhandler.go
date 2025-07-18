package logging

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"time"
)

// SessionHandler writes logs to a file in JSONL format.
type SessionHandler struct {
	file *os.File
}

// NewSessionHandler creates a new SessionHandler for the given file path.
func NewSessionHandler(path string) *SessionHandler {
	file, _ := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) // Error handling omitted for brevity
	return &SessionHandler{file: file}
}

// Enabled allows all log levels for session logs.
func (h *SessionHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

// Handle writes the log record as a JSONL entry with a timestamp.
func (h *SessionHandler) Handle(ctx context.Context, r slog.Record) error {
	logEntry := map[string]interface{}{
		"time":    time.Now().Format(time.RFC3339),
		"level":   r.Level.String(),
		"message": r.Message,
	}
	r.Attrs(func(a slog.Attr) bool {
		logEntry[a.Key] = a.Value.Any()
		return true
	})
	data, err := json.Marshal(logEntry)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = h.file.Write(data)
	return err
}

// WithAttrs returns the handler unchanged (attributes are included in Handle).
func (h *SessionHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

// WithGroup returns the handler unchanged (groups not needed for session logs).
func (h *SessionHandler) WithGroup(name string) slog.Handler {
	return h
}
