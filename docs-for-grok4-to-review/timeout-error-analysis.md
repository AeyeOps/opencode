# Timeout Error Analysis: "Failed General. Fail to generate. Title. Context. Deadline exceeded"

## Error Source
This error appears to be coming from the LLM provider (likely xAI/Grok based on context), not from OpenCode itself.

## Likely Causes

### 1. Request Timeout
The LLM provider took too long to generate a response and hit an internal timeout.

### 2. Context Length
The conversation might have too many tokens, causing processing delays.

### 3. Provider Overload
The LLM service might be experiencing high load.

## Current Timeout Configurations

### OpenCode Timeouts:
- MCP tools initialization: 30 seconds
- LSP client initialization: 30 seconds
- LSP client shutdown: 5 seconds
- Fetch tool HTTP requests: 30 seconds (configurable up to 120 seconds)

### Missing Timeouts:
- **LLM provider clients have NO explicit timeout configuration**
- This means they rely on the provider's default timeouts

## Potential Fixes

### 1. Add Configurable Timeouts to LLM Providers
Add timeout configuration to each provider client initialization:

```go
// For OpenAI-based providers (including xAI)
clientOptions := []option.RequestOption{
    option.WithHTTPClient(&http.Client{
        Timeout: 120 * time.Second, // 2 minutes
    }),
}
```

### 2. Implement Retry Logic with Backoff
The providers already have retry logic for rate limits (429 errors), but not for timeouts.

### 3. Add Context Timeout Wrapping
Wrap LLM calls with context timeouts:

```go
ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
defer cancel()
```

### 4. Session Management
Consider implementing:
- Automatic context pruning when nearing limits
- Warning users about large contexts
- Option to start fresh session

## Immediate Workarounds

1. **Start a new session** - Clear the conversation history
2. **Shorter prompts** - Break complex requests into smaller parts
3. **Switch models** - Some models handle long contexts better
4. **Wait and retry** - Provider load might be temporary

## Recommended Implementation

1. Add configurable timeout to provider options
2. Expose timeout configuration in settings
3. Add better error messages for timeout scenarios
4. Implement automatic retry with exponential backoff for timeout errors