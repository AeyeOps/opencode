# OpenCode Architecture Knowledge Base

## Overview
OpenCode is a terminal-based AI assistant for software development written in Go. It provides an interactive chat interface with multiple LLM provider support, code analysis, and LSP integration.

## Key Components

### 1. Configuration System (`internal/config/`)
- **Main Config**: `config.go` - Central configuration management
- **Config Structure**: 
  - Providers map: `map[models.ModelProvider]Provider`
  - Agents map: `map[AgentName]Agent` 
  - Each agent has: model ID, max tokens, reasoning effort
  - Provider has: API key, disabled flag

### 2. Model System (`internal/llm/models/`)
- **Model Registry**: Each provider has its own file defining models
  - `anthropic.go` - Claude models
  - `openai.go` - GPT models (including O1, O3, O4)
  - `gemini.go` - Google Gemini models
  - `groq.go` - Groq models
  - `azure.go` - Azure OpenAI models
  - `copilot.go` - GitHub Copilot models
  - `openrouter.go` - OpenRouter models
  - `vertexai.go` - Vertex AI models
  - `xai.go` - X.AI models
- **Central Registry**: `models.go` combines all models into `SupportedModels` map

### 3. Provider System (`internal/llm/provider/`)
- **Provider Interface**: `provider.go` defines common interface
- **Provider Implementations**: Each provider has client implementation
- **Factory Pattern**: `NewProvider()` creates appropriate provider based on model

### 4. TUI System (`internal/tui/`)
- **Model Selection Dialog**: `components/dialog/models.go`
  - Shows only enabled providers (those with API keys and not disabled)
  - Allows switching between providers with arrow keys
  - Filters models by selected provider

## Configuration Flow

1. **Load Config** (`config.Load()`)
   - Initialize empty providers map
   - Configure viper for config file reading
   - Set defaults based on environment variables
   - Read global config (~/.opencode.json)
   - Merge local config (.opencode.json)
   - Run `setProviderDefaults()` to set default models
   - Unmarshal config with viper
   - Validate configuration

2. **Provider Detection** (`setProviderDefaults()`)
   - Checks environment variables in priority order:
     1. GitHub Copilot (from GitHub config files)
     2. Anthropic (ANTHROPIC_API_KEY)
     3. OpenAI (OPENAI_API_KEY)
     4. Gemini (GEMINI_API_KEY)
     5. Groq (GROQ_API_KEY)
     6. OpenRouter (OPENROUTER_API_KEY)
     7. X.AI (XAI_API_KEY)
     8. AWS Bedrock (AWS credentials)
     9. Azure (AZURE_OPENAI_ENDPOINT)
     10. Vertex AI (Google Cloud credentials)

3. **Validation** (`Validate()`)
   - For each agent, validates:
     - Model exists in registry
     - Provider is configured
     - Provider has API key
     - Max tokens are valid
   - If validation fails, reverts to default model

## Provider Enable/Disable Logic

### How Providers Get Into Config:
1. **From Config File**: Explicitly defined in ~/.opencode.json
2. **From Environment**: Added during validation if model needs it
3. **Never Added**: If no API key found anywhere

### Why Only Some Providers Show:
- `getEnabledProviders()` only returns providers that:
  1. Exist in `cfg.Providers` map
  2. Have `Disabled: false`
  3. Have non-empty API key

### Current Issue Analysis:
The user has these providers in config file:
- anthropic ✓
- gemini ✓  
- openai ✓
- xai ✓

But model selection might only show OpenAI because:
1. The configured model "o3" is an OpenAI model
2. During validation, if other providers aren't needed for any agent, they might not be properly initialized

## Agent Types
- `coder`: Main coding assistant
- `summarizer`: For summarizing content
- `task`: For task management
- `title`: For generating titles (limited tokens)

## Key Insights

1. **Provider Population**: Providers only exist in the runtime config if they're either in the config file OR needed by a configured model and have env vars.

2. **Model Validation**: When an unsupported model is configured, the system falls back to defaults based on available providers.

3. **Dynamic Provider Addition**: During validation, if a model needs a provider not in config but has env vars, it gets added.

4. **TUI Provider List**: Only shows providers that are in the config AND enabled AND have API keys.