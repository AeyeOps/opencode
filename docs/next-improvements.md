# Next Improvements for OpenCode + Grok-4

## 1. Enhanced Tool Name Cleaning (Immediate)
The current repetition fix handles 2-3 repetitions, but Grok-4 showed 10+ repetitions. Add more aggressive cleaning:

```go
// In agent.go, before the existing repetition check
func cleanToolName(name string) string {
    // Handle extreme repetitions by counting occurrences
    knownTools := []string{"edit", "write", "view", "bash", "grep", "glob", "ls", "fetch", "patch", "sourcegraph", "agent"}
    
    for _, tool := range knownTools {
        if strings.Contains(name, tool) && strings.Count(name, tool) > 1 {
            return tool
        }
    }
    
    // Handle combined names
    if strings.Contains(name, "sourcegraph") && strings.Contains(name, "view") {
        return "sourcegraph" // Default to first tool
    }
    
    return name
}
```

## 2. Model-Specific Error Messages
When Grok-4 fails, provide specific guidance:

```go
if model == "grok-4" && strings.Contains(err.Error(), "Tool not found") {
    return fmt.Sprintf("Tool not found: %s\n\nGrok-4 tip: Use exact tool names from the list. Common issues:\n- Don't repeat names (use 'edit' not 'editeditedit')\n- Don't combine names (use 'sourcegraph' not 'sourcegraphview')\n- Check available tools with 'help'", toolName)
}
```

## 3. Tool Call Validation Layer
Add pre-flight validation before executing tools:

```go
type ToolCallValidator struct {
    model string
    stats map[string]int // Track error patterns
}

func (v *ToolCallValidator) Validate(call *message.ToolCall) error {
    // Check for repeated names
    if detectRepetition(call.Name) {
        call.Name = cleanToolName(call.Name)
        v.stats["repetition"]++
    }
    
    // Check for combined names
    if detectCombination(call.Name) {
        return fmt.Errorf("invalid tool name '%s': appears to combine multiple tools", call.Name)
    }
    
    return nil
}
```

## 4. Grok-4 Specific Features

### A. Retry with Guidance
After a tool error, inject a system message:
```go
if toolError && model == "grok-4" {
    // Add a temporary message to guide the model
    messages = append(messages, message.System{
        Content: "Remember: Use ONLY these exact tool names: bash, edit, view, grep, glob, ls, fetch, patch, write, sourcegraph, agent"
    })
}
```

### B. Tool Usage Statistics
Track and display tool call success/failure rates:
```go
type ModelStats struct {
    ToolCalls      int
    ToolErrors     int
    RepetitionErrs int
    CombinationErrs int
}

// Display in status bar or on exit
```

## 5. Live Prompt Testing Mode
Add a command to test prompts without executing tools:

```bash
opencode-grok4 --test-prompt
```

This would:
- Load the current prompt
- Show what tools would be called
- Validate tool names without execution
- Help iterate on prompt improvements

## 6. Fallback Strategies

### A. Auto-correction with Confirmation
```go
if tool == nil && attemptedTool := guessIntendedTool(toolCall.Name); attemptedTool != nil {
    // Ask user: "Did you mean 'edit'? (y/n)"
    if confirmed {
        toolCall.Name = attemptedTool.Name
        tool = attemptedTool
    }
}
```

### B. Pattern-based Recovery
```go
// If Grok-4 fails 3+ times with tools, offer alternatives:
"Grok-4 is having difficulty with tool calling. Options:
1. Switch to Claude/GPT-4 (recommended)
2. Use manual mode (paste commands)
3. Continue with auto-correction enabled"
```

## 7. Monitoring and Telemetry
Add optional telemetry to track Grok-4 issues:
- Tool error patterns
- Prompt effectiveness
- Success rates by prompt version

## 8. Documentation
Create a Grok-4 troubleshooting guide:
- Common errors and fixes
- Prompt engineering tips
- Known limitations
- Workaround strategies

## Priority Implementation Order

1. **Now**: Test the externalized prompt to see if it helps
2. **Next**: Add aggressive tool name cleaning
3. **Then**: Model-specific error messages
4. **Later**: Statistics and monitoring
5. **Future**: Advanced features like auto-correction

## Testing Strategy

1. Create a test suite specifically for Grok-4:
   ```bash
   # Test file with known problematic patterns
   echo "Test tool calling" > test_grok4.txt
   opencode-grok4 "search for examples then view a file"
   ```

2. Track which prompt modifications work best

3. Share findings with xAI team for model improvements