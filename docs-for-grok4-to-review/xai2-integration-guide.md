# XAI2 Provider Integration Guide

## Overview
The XAI2 provider is a standalone implementation of xAI models for OpenCode, designed to work alongside the existing XAI provider without conflicts.

## Model Specifications

### Grok-4 (Flagship Model)
- **Model IDs**: `xai2.grok-4`, `xai2.grok-4-0709`
- **Context Window**: 256,000 tokens
- **Default Max Tokens**: 20,000 tokens
- **Features**: 
  - Reasoning support
  - Function calling (tool use)
  - Cached input pricing available
- **Pricing**:
  - Input: $3.00 per 1M tokens
  - Input (cached): $0.75 per 1M tokens
  - Output: $15.00 per 1M tokens

### Grok-3 Models
- **Context Window**: 131,072 tokens
- **Default Max Tokens**: 20,000 tokens
- **Variants**:
  - `xai2.grok-3` / `xai2.grok-3-beta`: Standard model ($3/$15 per 1M)
  - `xai2.grok-3-mini` / `xai2.grok-3-mini-beta`: Lightweight ($0.30/$0.50 per 1M)
  - `xai2.grok-3-fast` / `xai2.grok-3-fast-beta`: Fast response ($5/$25 per 1M)
  - `xai2.grok-3-mini-fast` / `xai2.grok-3-mini-fast-beta`: Mini fast ($0.60/$4 per 1M)

### Grok-2 Vision
- **Model IDs**: `xai2.grok-2-vision`, `xai2.grok-2-vision-1212`
- **Context Window**: 32,768 tokens
- **Default Max Tokens**: 8,000 tokens
- **Features**: Multimodal (text + vision)
- **Pricing**: $2/$10 per 1M tokens

## Configuration

### Config File Location
OpenCode looks for configuration in these locations (in order):
1. `~/.opencode.json`
2. `$XDG_CONFIG_HOME/opencode/.opencode.json`
3. `$HOME/.config/opencode/.opencode.json`

### Sample Configuration
```json
{
  "providers": {
    "xai2": {
      "apiKey": "xai-YOUR-API-KEY-HERE",
      "disabled": false
    }
  },
  "agents": {
    "coder": {
      "model": "xai2.grok-4",
      "maxTokens": 20000,
      "reasoningEffort": ""
    }
  }
}
```

### Agent Configuration Options
- **model**: The model ID (e.g., `xai2.grok-4`)
- **maxTokens**: Maximum output tokens (up to model's DefaultMaxTokens)
- **reasoningEffort**: Only applicable to OpenAI models, not used for xAI

## Testing the Integration

### Build OpenCode
```bash
go build -o opencode
```

### Test Non-Interactive Mode
```bash
# Simple test with debug output
./opencode -p "What is 2+2?" -d

# Quiet mode (no spinner)
./opencode -p "Explain quantum computing" -q
```

### Test Interactive Mode
```bash
./opencode
# Press Ctrl+O to open model selection dialog
# Verify xai2.grok-4 appears in the list
```

### Command-Line Flags
- `-h, --help`: Show help
- `-v, --version`: Show version
- `-d, --debug`: Enable debug mode (shows provider/model info)
- `-c, --cwd`: Set working directory
- `-p, --prompt`: Non-interactive mode with prompt
- `-f, --output-format`: Output format (text, json)
- `-q, --quiet`: Hide spinner in non-interactive mode

## Implementation Details

### Model Name Collision Prevention
All XAI2 models use the `xai2.` prefix to prevent conflicts with the original XAI provider:
- XAI provider: `grok-4`, `grok-3`, etc.
- XAI2 provider: `xai2.grok-4`, `xai2.grok-3`, etc.

### Provider Registration
The XAI2 provider is registered in:
- `internal/llm/models/models.go`: Added to ProviderPopularity (priority 11)
- `internal/llm/provider/provider.go`: Case handled in NewProvider()
- `internal/llm/models/models.go`: XAI2Models copied to SupportedModels in init()

### Error Handling
The XAI2 client includes specific error handling for:
- Rate limiting patterns
- Quota exceeded errors
- Content errors (when xAI returns errors as message content)
- Retry logic with exponential backoff (up to 8 retries)

## Troubleshooting

### Model Not Appearing
1. Verify the provider is enabled in config (`"disabled": false`)
2. Check API key is set correctly
3. Run with `-d` flag to see debug output
4. Ensure XAI2 models are in SupportedModels map

### API Errors
- Check for rate limiting messages in response content
- Verify API key has sufficient quota
- Monitor for specific error patterns (defined in xai2ErrorPatterns)