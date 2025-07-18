# OpenCode Code Structure

## Directory Layout

```
opencode/
├── cmd/                    # Command line interface
│   └── root.go            # Main command setup
├── internal/              # Internal packages
│   ├── app/              # Application core
│   ├── config/           # Configuration management
│   │   ├── config.go     # Main config logic
│   │   └── init.go       # Initialization helpers
│   ├── db/               # Database layer
│   │   └── models.go     # DB models
│   ├── llm/              # LLM integration
│   │   ├── models/       # Model definitions
│   │   │   ├── models.go # Model registry
│   │   │   ├── anthropic.go
│   │   │   ├── openai.go
│   │   │   ├── gemini.go
│   │   │   ├── groq.go
│   │   │   ├── azure.go
│   │   │   ├── copilot.go
│   │   │   ├── openrouter.go
│   │   │   ├── vertexai.go
│   │   │   ├── xai.go
│   │   │   └── local.go
│   │   ├── provider/     # Provider implementations
│   │   │   ├── provider.go    # Provider interface
│   │   │   ├── anthropic.go
│   │   │   ├── openai.go
│   │   │   ├── gemini.go
│   │   │   ├── azure.go
│   │   │   ├── bedrock.go
│   │   │   ├── copilot.go
│   │   │   └── vertexai.go
│   │   ├── agent/        # Agent logic
│   │   └── tools/        # Tool definitions
│   ├── tui/              # Terminal UI
│   │   ├── components/   # UI components
│   │   │   └── dialog/
│   │   │       └── models.go  # Model selection dialog
│   │   └── theme/        # UI themes
│   ├── logging/          # Logging utilities
│   ├── message/          # Message handling
│   └── pubsub/           # Pub/sub system
├── scripts/              # Build/utility scripts
├── main.go              # Entry point
├── go.mod               # Go modules
├── .opencode.json       # Local config
└── opencode-schema.json # Config schema
```

## Key Interfaces

### Provider Interface
```go
type Provider interface {
    SendMessages(ctx, messages, tools) (*ProviderResponse, error)
    StreamResponse(ctx, messages, tools) <-chan ProviderEvent
    Model() models.Model
}
```

### Model Structure
```go
type Model struct {
    ID                  ModelID
    Name                string
    Provider            ModelProvider
    APIModel            string
    CostPer1MIn         float64
    CostPer1MOut        float64
    ContextWindow       int64
    DefaultMaxTokens    int64
    CanReason           bool
    SupportsAttachments bool
}
```

### Config Structure
```go
type Config struct {
    Data         Data
    WorkingDir   string
    MCPServers   map[string]MCPServer
    Providers    map[models.ModelProvider]Provider
    LSP          map[string]LSPConfig
    Agents       map[AgentName]Agent
    Debug        bool
    TUI          TUIConfig
    Shell        ShellConfig
    AutoCompact  bool
}
```

## Model Registration

Models are registered in a two-step process:

1. **Provider-specific maps** (e.g., `AnthropicModels`, `OpenAIModels`)
2. **Central registry** (`SupportedModels`) via `init()` function:
   ```go
   func init() {
       maps.Copy(SupportedModels, AnthropicModels)
       maps.Copy(SupportedModels, OpenAIModels)
       // ... etc
   }
   ```

## Provider Factory

The `NewProvider()` function creates providers based on the provider type:
```go
switch providerName {
case models.ProviderAnthropic:
    return &baseProvider[AnthropicClient]{...}
case models.ProviderOpenAI:
    return &baseProvider[OpenAIClient]{...}
// ... etc
}
```

Some providers (GROQ, OpenRouter, XAI) use the OpenAI client with different base URLs.

## Configuration Priority

1. **Environment Variables** (highest priority)
2. **Local Config** (.opencode.json in working directory)
3. **Global Config** (~/.opencode.json)
4. **Defaults** (lowest priority)

## Agent Configuration

Each agent can be configured with:
- `model`: The model ID to use
- `maxTokens`: Maximum tokens for responses
- `reasoningEffort`: For models that support reasoning (low/medium/high)

Default agents:
- `coder`: Main assistant for coding tasks
- `summarizer`: For summarizing content
- `task`: For task management
- `title`: For generating titles (limited tokens)