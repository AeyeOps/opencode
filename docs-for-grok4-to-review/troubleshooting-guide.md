# OpenCode Troubleshooting Guide

## Common Issues

### 1. "No such device or address" Error
**Error**: `could not open a new TTY: open /dev/tty: no such device or address`
**Cause**: OpenCode requires a TTY (terminal) to run the TUI interface
**Solution**: Run in a proper terminal, not in a pipe or background process

### 2. Only One Provider Shows in Model Selection
**Symptoms**: Despite having multiple providers configured, only one shows up
**Possible Causes**:
1. Providers are disabled during validation
2. API keys are invalid or empty
3. Config file format issues
4. Provider not needed by any agent

**Debugging Steps**:
1. Check ~/.opencode.json for provider configuration
2. Verify API keys are valid
3. Run with `-d` flag for debug logging
4. Check if providers have `"disabled": false`

### 3. Model Not Found
**Error**: "unsupported model configured, reverting to default"
**Cause**: Configured model ID doesn't exist in registry
**Solution**: Check available models in `internal/llm/models/` files

### 4. Provider Not Configured
**Error**: "provider not configured for model"
**Cause**: Model's provider lacks API key or is disabled
**Solution**: Add provider API key to config or environment

## Debug Commands

```bash
# Run with debug logging
opencode -d

# Run with specific working directory
opencode -c /path/to/project

# Non-interactive mode with prompt
opencode -p "Your prompt here"

# Check version
opencode -v
```

## Configuration Debugging

### 1. Check Config Files
```bash
# Global config
cat ~/.opencode.json

# Local config
cat .opencode.json

# Check environment variables
env | grep -E "(ANTHROPIC|OPENAI|GEMINI|GROQ|OPENROUTER|XAI)_API_KEY"
```

### 2. Config File Structure
```json
{
  "providers": {
    "anthropic": {
      "apiKey": "sk-ant-...",
      "disabled": false
    },
    "openai": {
      "apiKey": "sk-proj-...",
      "disabled": false
    }
  },
  "agents": {
    "coder": {
      "model": "claude-3.7-sonnet",
      "maxTokens": 5000
    }
  }
}
```

### 3. Provider Priority Order
1. GitHub Copilot (if available)
2. Anthropic
3. OpenAI
4. Gemini
5. Groq
6. OpenRouter
7. X.AI
8. AWS Bedrock
9. Azure
10. Vertex AI

## API Key Sources

### Environment Variables
- `ANTHROPIC_API_KEY`
- `OPENAI_API_KEY`
- `GEMINI_API_KEY`
- `GROQ_API_KEY`
- `OPENROUTER_API_KEY`
- `XAI_API_KEY`
- `AZURE_OPENAI_API_KEY` + `AZURE_OPENAI_ENDPOINT`
- `GITHUB_TOKEN` (for Copilot)
- AWS credentials (for Bedrock)
- Google Cloud credentials (for Vertex AI)

### GitHub Copilot Token
Automatically loaded from:
- `$XDG_CONFIG_HOME/github-copilot/hosts.json`
- `$XDG_CONFIG_HOME/github-copilot/apps.json`
- Windows: `%LOCALAPPDATA%/github-copilot/`

## Model Selection Logic

1. **Current Model**: Determined by `agents.coder.model` in config
2. **Available Providers**: Only those with API keys and not disabled
3. **Provider Models**: Filtered by selected provider
4. **Navigation**: Arrow keys to switch providers, Enter to select

## Validation Flow

1. **Model Exists**: Check if model ID is in `SupportedModels`
2. **Provider Check**: Verify provider has API key
3. **Token Validation**: Ensure max tokens are valid
4. **Fallback**: If validation fails, revert to default based on available providers

## Common Config Mistakes

1. **Wrong Model ID**: Using "gpt-4" instead of "gpt-4.1"
2. **Missing API Key**: Provider in config but no API key
3. **Disabled Provider**: `"disabled": true` in config
4. **Invalid JSON**: Syntax errors in config file
5. **Wrong Provider**: Model ID doesn't match provider

## Log File Locations

When `OPENCODE_DEV_DEBUG=true`:
- Debug log: `~/.opencode/debug.log`
- Messages: `~/.opencode/messages/`

## Building from Source

```bash
# Clone repository
git clone https://github.com/opencode-ai/opencode
cd opencode

# Build
go build -o opencode .

# Install
./install
```

## Resetting Configuration

```bash
# Backup current config
cp ~/.opencode.json ~/.opencode.json.bak

# Remove config
rm ~/.opencode.json

# Remove initialization flag
rm ~/.opencode/init

# Start fresh
opencode
```