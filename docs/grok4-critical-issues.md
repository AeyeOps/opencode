# Critical Issues with Grok-4 in OpenCode

## Observed Problems

1. **Extreme Tool Name Repetition**
   - `editeditediteditediteditediteditedit` (edit repeated 10 times!)
   - Much worse than other models' occasional doubling

2. **Tool Name Hallucination**
   - `sourcegraphview` - combining two separate tools
   - Shows fundamental misunderstanding of available tools

3. **Patch Format Errors**
   - Invalid context in patches
   - Malformed patch syntax
   - Duplicate lines in context

4. **Persistent Failure Loop**
   - Keeps trying the same failed operation
   - Doesn't learn from "Tool not found" errors
   - No error recovery

## Root Causes

1. **Model Architecture**: Grok-4 may have issues with:
   - Tokenization of tool names
   - Understanding tool boundaries
   - Following structured output formats

2. **System Prompt Ineffectiveness**: Current prompt doesn't prevent:
   - Tool name corruption
   - Invalid tool combinations
   - Repeated failed attempts

## Immediate Workarounds

### 1. Disable Grok-4
Until fixes are implemented, consider removing Grok-4 from available models:
```go
// In models/xai.go, comment out Grok-4
```

### 2. Enhanced Tool Name Cleaning
Add more aggressive cleaning in agent.go:
```go
// Clean extreme repetitions first
if strings.Count(toolCall.Name, "edit") > 3 {
    toolCall.Name = "edit"
} else if strings.Count(toolCall.Name, "write") > 3 {
    toolCall.Name = "write"
}
// Then existing repetition fix...
```

### 3. Tool Combination Detection
```go
// Detect combined tool names
if strings.Contains(toolCall.Name, "sourcegraph") && strings.Contains(toolCall.Name, "view") {
    // Default to first tool found
    if strings.Index(toolCall.Name, "sourcegraph") < strings.Index(toolCall.Name, "view") {
        toolCall.Name = "sourcegraph"
    } else {
        toolCall.Name = "view"
    }
}
```

## Long-term Solutions

### 1. Model-Specific Validation
Create a validation layer for Grok-4:
```go
type ModelValidator interface {
    ValidateToolCall(model string, toolCall *message.ToolCall) error
    CleanToolCall(model string, toolCall *message.ToolCall)
}
```

### 2. Retry with Different Model
After N failures with Grok-4, offer to switch models:
```go
if consecutiveToolErrors > 3 && model == "grok-4" {
    return "Multiple tool errors detected. Consider switching to a different model."
}
```

### 3. Pre-flight Tool Validation
Validate tool calls before execution:
```go
func (a *Agent) validateToolCall(toolCall message.ToolCall) error {
    // Check tool exists
    // Validate tool name format
    // Check for known problematic patterns
}
```

## Testing Recommendations

1. Create test suite specifically for Grok-4 tool calling
2. Test with various prompt styles
3. Monitor tool error rates by model
4. Consider A/B testing different prompts

## User Communication

When Grok-4 fails repeatedly:
```
"Grok-4 is having difficulty with tool formatting. This is a known issue. 
Would you like to:
1. Switch to another model (recommended)
2. Continue with manual workarounds
3. Report this for improvement"
```

## Priority Actions

1. **Immediate**: Add aggressive tool name cleaning
2. **Short-term**: Model-specific error handling
3. **Medium-term**: Enhanced prompts for Grok-4
4. **Long-term**: Work with xAI to improve model behavior