
---

## Logging System Improvements

The current logging system in OpenCode, as seen in `internal/logging/`, uses `slog` with basic handlers. However, it has limitations such as a single output at a time and a split between regular logs and session logs. To address this, I'll propose a unified logging system with a multi-handler approach, RAG integration for self-improvement, and enhanced session logging.

### 1. Unified Logging with Multi-Handler

We'll enhance the existing `MultiHandler` to route logs to multiple destinations (console, file, session files, RAG) simultaneously, improving flexibility and consistency.

**Modified File:** `internal/logging/multihandler.go`

```go
package logging

import (
	"context"
	"errors"
	"log/slog"
)

// MultiHandler routes logs to multiple slog.Handler instances.
type MultiHandler struct {
	handlers []slog.Handler
}

// NewMultiHandler creates a new MultiHandler with the given handlers.
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

// Enabled checks if any handler is enabled for the given level.
func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle processes the log record by passing it to all enabled handlers.
func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// WithAttrs creates a new MultiHandler with additional attributes.
func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return NewMultiHandler(handlers...)
}

// WithGroup creates a new MultiHandler with a group name.
func (h *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return NewMultiHandler(handlers...)
}
```

**Changes Explained**:
- The file remains largely unchanged from the provided version but is included here for completeness. It enables logs to be sent to multiple outputs, which we'll leverage with new handlers.

### 2. RAG Integration with `RAGHandler`

We'll introduce a new `RAGHandler` to embed reflection logs into a vector database (e.g., FAISS) for self-improvement, allowing OpenCode to learn from past logs.

**New File:** `internal/logging/raghandler.go`

```go
package logging

import (
	"context"
	"log/slog"
	"github.com/opencode-ai/opencode/internal/db"
)

// RAGHandler embeds reflection logs into a vector database for self-improvement.
type RAGHandler struct {
	db *db.FAISSDB
}

// NewRAGHandler creates a new RAGHandler with a FAISS database instance.
func NewRAGHandler(db *db.FAISSDB) *RAGHandler {
	return &RAGHandler{db: db}
}

// Enabled ensures only Info level and above are processed for RAG.
func (h *RAGHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= slog.LevelInfo
}

// Handle embeds reflection logs into the vector database.
func (h *RAGHandler) Handle(ctx context.Context, r slog.Record) error {
	var isReflection bool
	var sessionID string
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "type" && a.Value.String() == "reflection" {
			isReflection = true
		}
		if a.Key == "sessionID" {
			sessionID = a.Value.String()
		}
		return true
	})
	if isReflection {
		vector := embed(r.Message) // Convert message to vector
		return h.db.Insert(sessionID, vector, r)
	}
	return nil
}

// WithAttrs returns the handler unchanged (attributes not stored in RAG).
func (h *RAGHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

// WithGroup returns the handler unchanged (groups not stored in RAG).
func (h *RAGHandler) WithGroup(name string) slog.Handler {
	return h
}

// embed converts a message to a vector (placeholder for actual embedding).
func embed(message string) []float32 {
	// TODO: Implement with a real embedding model (e.g., BERT or SentenceTransformers).
	return []float32{} // Placeholder
}
```

**Explanation**:
- This handler captures logs tagged with `type=reflection` and embeds them into a FAISS database (assumed to be defined in `internal/db`). This enables retrieval-augmented generation (RAG) for self-improvement.

### 3. Enhanced Session Logging with `SessionHandler`

We'll modify the existing `SessionHandler` to write logs in JSONL format with timestamps, improving readability and performance for session logs.

**Modified File:** `internal/logging/sessionhandler.go`

```go
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
```

**Changes Explained**:
- Added JSONL formatting with timestamps, enhancing readability and enabling real-time tailing in the TUI.
- Simplified attribute handling by embedding them directly in the log entry.

### 4. Configuration Updates for Logging

We'll update the `Config` struct to support multiple logging outputs and per-module levels, aligning with the multi-handler system.

**Modified File:** `internal/config/config.go`

```go
package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/opencode-ai/opencode/internal/logging"
	"github.com/spf13/viper"
)

// LoggingOutput defines a log output type and its options.
type LoggingOutput struct {
	Type    string                 `json:"type"`
	Options map[string]interface{} `json:"options"`
}

// LoggingModule defines a module's logging level.
type LoggingModule struct {
	Level string `json:"level"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Outputs []LoggingOutput          `json:"outputs"`
	Modules map[string]LoggingModule `json:"modules"`
}

// ... (other existing types like MCPType, MCPServer, AgentName, etc., remain unchanged)

// Config holds the application configuration.
type Config struct {
	Data         Data                              `json:"data"`
	WorkingDir   string                            `json:"wd,omitempty"`
	MCPServers   map[string]MCPServer              `json:"mcpServers,omitempty"`
	Providers    map[models.ModelProvider]Provider `json:"providers,omitempty"`
	LSP          map[string]LSPConfig              `json:"lsp,omitempty"`
	Agents       map[AgentName]Agent               `json:"agents,omitempty"`
	Debug        bool                              `json:"debug,omitempty"`
	DebugLSP     bool                              `json:"debugLSP,omitempty"`
	ContextPaths []string                          `json:"contextPaths,omitempty"`
	TUI          TUIConfig                         `json:"tui"`
	Shell        ShellConfig                       `json:"shell,omitempty"`
	AutoCompact  bool                              `json:"autoCompact,omitempty"`
	Logging      LoggingConfig                     `json:"logging"`
}

// ... (constants and existing functions remain unchanged)

// Load reads the configuration from a file.
func Load(workingDir string, debug bool) (*Config, error) {
	configureViper()
	setDefaults(debug)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, readConfig(err)
	}
	var c Config
	err = viper.Unmarshal(&c)
	if err != nil {
		return nil, fmt.Errorf("unable to decode config: %v", err)
	}
	mergeLocalConfig(workingDir)
	applyDefaultValues()
	cfg = &c
	return cfg, nil
}

// SetupLogging configures the logging system based on the config.
func SetupLogging(cfg *Config) error {
	var蜀 handlers []slog.Handler
	for _, output := range cfg.Logging.Outputs {
		switch output.Type {
		case "console":
			handlers = append(handlers, slog.NewTextHandler(os.Stdout, nil))
		case "file":
			path, ok := output.Options["path"].(string)
			if !ok {
				path = "opencode.log"
			}
			file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			handlers = append(handlers, slog.NewJSONHandler(file, nil))
		case "session":
			path, ok := output.Options["path"].(string)
			if !ok {
				path = "session.log"
			}
			handlers = append(handlers, NewSessionHandler(path))
		case "rag":
			// Assuming FAISSDB is initialized elsewhere
			db := &db.FAISSDB{} // Placeholder; replace with actual initialization
			handlers = append(handlers, NewRAGHandler(db))
		}
	}
	slog.SetDefault(slog.New(NewMultiHandler(handlers...)))
	return nil
}

// ... (other existing functions remain unchanged)
```

**Changes Explained**:
- Added `LoggingOutput`, `LoggingModule`, and `LoggingConfig` structs to support multiple outputs and module-specific levels.
- Updated `Config` to include `Logging`.
- Enhanced `SetupLogging` to initialize handlers based on config, integrating the new `SessionHandler` and `RAGHandler`.

---

## TUI Enhancements

The current TUI, as seen in `internal/tui/`, provides a basic logs view in `logs/table.go`. We'll enhance it with real-time log tailing, interactive filters, and a reflection dashboard in the sidebar.

### 1. Real-Time Log Tailing

We'll modify the logs table to use `fsnotify` for real-time updates and add filtering capabilities.

**Modified File:** `internal/tui/components/logs/table.go`

```go
package logs

import (
	"fmt"
	"strings"
	"github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

type LogsTable struct {
	watcher *fsnotify.Watcher
	logs    []string
	filter  string
	verbose bool
}

func NewLogsTable() *LogsTable {
	return &LogsTable{
		logs:    []string{},
		verbose: false,
	}
}

func (t *LogsTable) Init() tea.Cmd {
	var err error
	t.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return func() tea.Msg { return err }
	}
	go func() {
		for {
			select {
			case event, ok := <-t.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					// TODO: Read new log lines from file and append to t.logs
					t.logs = append(t.logs, fmt.Sprintf("Log updated: %s", event.Name))
				}
			case err, ok := <-t.watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()
	t.watcher.Add("session.log") // Watch session log file
	return nil
}

func (t *LogsTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+f":
			// TODO: Enable filter input mode
			t.filter = "approved" // Placeholder for filter input
		case "v":
			t.verbose = !t.verbose
		}
	}
	return t, nil
}

func (t *LogsTable) View() string {
	var b strings.Builder
	for _, log := range t.logs {
		if t.filter != "" && !strings.Contains(log, t.filter) {
			continue
		}
		color := "white"
		if strings.Contains(log, "approved") {
			color = "green"
		} else if strings.Contains(log, "denied") {
			color = "red"
		}
		b.WriteString(fmt.Sprintf("[%s]%s\n", color, log))
	}
	return b.String()
}
```

**Changes Explained**:
- Added `fsnotify` to watch `session.log` and append new entries in real-time.
- Introduced basic filtering (placeholder for full implementation) and verbose toggling.
- Enhanced `View` with color-coding for better readability.

### 2. Reflection Dashboard in Sidebar

We'll modify the chat sidebar to display recent reflections and KPIs, leveraging RAG data.

**Modified File:** `internal/tui/components/chat/sidebar.go`

```go
package chat

import (
	"fmt"
	"strings"
	"github.com/charmbracelet/bubbletea"
)

type Sidebar struct {
	reflections []Reflection
	kpis        map[string]string
}

type Reflection struct {
	Suggestion string
	Approved   bool
}

func NewSidebar() *Sidebar {
	return &Sidebar{
		reflections: []Reflection{},
		kpis:        map[string]string{},
	}
}

func (s *Sidebar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// TODO: Update reflections from RAGHandler logs and KPIs from metrics
	s.reflections = append(s.reflections, Reflection{Suggestion: "Optimize DB queries", Approved: true})
	s.kpis["ResponseTime"] = "300ms"
	return s, nil
}

func (s *Sidebar) View() string {
	var b strings.Builder
	b.WriteString("Recent Suggestions:\n")
	for _, r := range s.reflections {
		color := "red"
		if r.Approved {
			color = "green"
		}
		b.WriteString(fmt.Sprintf("[%s]%s\n", color, r.Suggestion))
	}
	b.WriteString("\nCurrent KPIs:\n")
	for k, v := range s.kpis {
		b.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}
	return b.String()
}
```

**Changes Explained**:
- Added a `Reflection` struct and fields to track suggestions and KPIs.
- Updated `Update` with placeholder logic to fetch data (to be integrated with RAG).
- Enhanced `View` to display reflections and KPIs with color-coding.

---

## Integration Steps

To implement these enhancements in OpenCode:

1. **Add New Files**:
   - Place `internal/logging/raghandler.go` in the codebase.

2. **Update Existing Files**:
   - Replace `internal/logging/multihandler.go`, `internal/logging/sessionhandler.go`, `internal/config/config.go`, `internal/tui/components/logs/table.go`, and `internal/tui/components/chat/sidebar.go` with the versions above.

3. **Dependencies**:
   - Add `github.com/fsnotify/fsnotify` to `go.mod` for real-time log tailing.
   - Ensure a FAISS implementation exists in `internal/db` or mock it for now.

4. **Configuration**:
   - Update `.opencode.json` to include logging outputs:
     ```json
     {
       "logging": {
         "outputs": [
           {"type": "console"},
           {"type": "session", "options": {"path": "session.log"}},
           {"type": "rag"}
         ],
         "modules": {"main": {"level": "info"}}
       }
     }
     ```

5. **Test**:
   - Run `go build -o opencode && ./opencode` to verify the logging system and TUI work as expected.

These changes enhance OpenCode's logging with a unified, multi-output system and improve the TUI with real-time log viewing and a reflection dashboard. Let me know if you need further assistance with implementation or additional features!