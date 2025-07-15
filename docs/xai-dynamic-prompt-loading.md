# XAI Dynamic Prompt Loading

## Overview

Both XAI and XAI2 providers support dynamic loading of system prompts from external files. This allows you to modify the Grok-4 system prompt without recompiling OpenCode.

## How It Works

1. **On each request**, the XAI/XAI2 provider checks for a prompt file named `grok4-system-prompt.md`
2. If found, the content of this file is used as the system message
3. This overrides any system message configured in the provider options

## File Search Path

The prompt file is searched in the same locations as `.opencode.json`:

1. `$HOME/.opencode/grok4-system-prompt.md`
2. `$XDG_CONFIG_HOME/opencode/grok4-system-prompt.md` (if XDG_CONFIG_HOME is set)
3. `$HOME/.config/opencode/grok4-system-prompt.md`

The first file found in this order is used.

## Implementation Details

### XAI Provider (`xai.go`)
```go
// Override convertMessages to reload Grok prompt from file on each request
func (x *xaiClient) convertMessages(messages []message.Message) []openai.ChatCompletionMessageParamUnion {
    // Reload the system message from file for Grok models
    if externalPrompt := x.loadExternalGrokPrompt(); externalPrompt != "" {
        x.providerOptions.systemMessage = externalPrompt
    }
    
    // Call the parent implementation
    return x.openaiClient.convertMessages(messages)
}
```

### XAI2 Provider (`xai2.go`)
```go
// Override convertMessages to reload Grok prompt from file on each request
func (x *xai2Client) convertMessages(messages []message.Message) []openai.ChatCompletionMessageParamUnion {
    // Reload the system message from file for Grok models
    if externalPrompt := x.loadExternalGrokPrompt(); externalPrompt != "" {
        x.providerOptions.systemMessage = externalPrompt
    }
    
    // Prepend system message if it exists
    openaiMessages := []openai.ChatCompletionMessageParamUnion{}
    if x.providerOptions.systemMessage != "" {
        openaiMessages = append(openaiMessages, openai.SystemMessage(x.providerOptions.systemMessage))
    }
    
    // Convert the rest of the messages
    for _, msg := range messages {
        openaiMessages = append(openaiMessages, x.convertMessage(msg))
    }
    return openaiMessages
}
```

## Key Differences

- **XAI**: Inherits from `openaiClient` and calls the parent's `convertMessages` method
- **XAI2**: Standalone implementation that manually prepends the system message

## Benefits

1. **Hot reload**: Change prompts without restarting OpenCode
2. **Experimentation**: Easy to test different prompts
3. **Version control**: Keep prompt files in git alongside your code
4. **Environment-specific**: Different prompts for different environments

## Example Usage

1. Create a prompt file:
```bash
mkdir -p ~/.config/opencode
cat > ~/.config/opencode/grok4-system-prompt.md << 'EOF'
You are Grok-4, an advanced AI assistant...
# Your custom prompt here
EOF
```

2. The prompt will be automatically loaded on the next request to XAI/XAI2

## Troubleshooting

- If the prompt isn't loading, check file permissions
- Ensure the file is in one of the search paths
- The file must be named exactly `grok4-system-prompt.md`
- Check OpenCode logs in debug mode to see which paths are being checked