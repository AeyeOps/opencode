# xAI Grok-4 System Prompt Recommendations

## Current Issues with Grok-4
1. **Invalid tool names**: Generating non-existent tools like "sourcegraphview" 
2. **Tool confusion**: Possibly merging multiple tool names
3. **Empty message errors**: Though this is already fixed in code

## Recommended Updates for Grok-4 Prompt

### 1. Add Explicit Tool List
Add a section explicitly listing available tools to prevent hallucination:

```
# Available Tools
You have access to exactly these tools (use exact names):
- bash: Execute shell commands
- edit: Modify existing files  
- fetch: Download content from URLs
- glob: Find files by pattern
- grep: Search file contents
- ls: List directory contents
- sourcegraph: Search code across public repositories
- view: View file contents
- patch: Apply patches to files
- write: Create new files
- agent: Delegate complex searches to sub-agent

IMPORTANT: Only use the exact tool names listed above. Do not combine or modify tool names.
```

### 2. Add Tool Usage Examples
Add concrete examples to guide proper tool usage:

```
# Tool Usage Examples
<example>
user: search for Whisper.net examples
assistant: I'll search for Whisper.net examples using Sourcegraph.
[Uses tool: sourcegraph with query "lang:csharp Whisper.net"]
</example>

<example>
user: view the project file
assistant: [Uses tool: view with file_path "/path/to/file.csproj"]
</example>
```

### 3. Add Explicit Tool Error Handling
```
# Tool Errors
If you receive "Tool not found" error:
1. Check you're using the exact tool name from the Available Tools list
2. Do not combine tool names (e.g., "sourcegraphview" is incorrect)
3. Use tools separately: first "sourcegraph" to search, then "view" to read files
```

### 4. Model-Specific Implementation
Instead of one xAI prompt for all models, implement model-specific prompts:

```go
func CoderPrompt(provider models.ModelProvider, model string) string {
    basePrompt := baseAnthropicCoderPrompt
    switch provider {
    case models.ProviderOpenAI:
        basePrompt = baseOpenAICoderPrompt
    case models.ProviderXAI:
        // Model-specific prompts for xAI
        if strings.Contains(model, "grok-4") {
            basePrompt = baseGrok4CoderPrompt
        } else {
            basePrompt = baseXAICoderPrompt
        }
    }
    // ...
}
```

### 5. Grok-4 Specific Prompt Structure
```
const baseGrok4CoderPrompt = `You are OpenCode, an interactive CLI tool powered by xAI's Grok-4 model.

CRITICAL: Tool Usage Rules
1. Use ONLY the exact tool names provided in the tool list
2. NEVER combine tool names (e.g., "sourcegraphview" does not exist)
3. Tools must be called with their exact names as separate operations

[Include existing prompt content with above additions...]
```

## Implementation Steps

1. Update `prompt/coder.go` to check for specific model versions
2. Create separate prompt constants for Grok-4
3. Add explicit tool list and examples
4. Test with various tool-calling scenarios

## Alternative: Runtime Tool Validation
If prompt updates don't fully resolve the issue, consider adding runtime validation:

```go
// In agent.go, before "Tool not found" error
if strings.Contains(toolCall.Name, "view") && strings.Contains(toolCall.Name, "sourcegraph") {
    // Likely meant to use these tools separately
    toolResults[i] = message.ToolResult{
        ToolCallID: toolCall.ID,
        Content:    fmt.Sprintf("Tool not found: %s. Did you mean to use 'sourcegraph' and 'view' as separate tools?", toolCall.Name),
        IsError:    true,
    }
    continue
}
```

This would provide more helpful error messages when models combine tool names.