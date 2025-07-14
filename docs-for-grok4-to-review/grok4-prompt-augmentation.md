# Grok-4 Prompt Augmentation

Add this section to your existing prompt:

## CRITICAL TOOL CALLING ERRORS - IMMEDIATE CORRECTION REQUIRED

### YOU ARE MAKING THESE ERRORS:
1. **lsglobglobglobglobglobglobglobglob** - This is WRONG. You are combining and repeating tool names.
2. **lsviewview** - This is WRONG. You are combining ls and view.
3. **editeditediteditediteditediteditedit** - This is WRONG. You are repeating edit.
4. **sourcegraphview** - This is WRONG. You are combining sourcegraph and view.

### CORRECT APPROACH - SEQUENTIAL TOOL CALLS:
When you need to do multiple operations, make SEPARATE tool calls:

```
Step 1: ls {"path": "/tmp/opencode"}
Step 2: glob {"pattern": "**/*.go"}
Step 3: view {"file_path": "/tmp/opencode/main.go"}
```

### TOOL CALLING ALGORITHM:
1. STOP before making a tool call
2. IDENTIFY the single operation you need
3. SELECT one tool from the available list
4. MAKE that single tool call
5. WAIT for the result
6. REPEAT for next operation

### SELF-CHECK BEFORE EVERY TOOL CALL:
Ask yourself:
- Am I using a tool name from the exact list provided?
- Is this tool name repeated (like editeditedit)?
- Is this tool name combined (like lsglob)?
- Have I used this exact tool name successfully before?

If ANY answer is "no" or "yes" to repetition/combination - STOP and correct it.

### PATTERN RECOGNITION:
- If you're thinking "list and search" - Use `ls` THEN `glob` (two calls)
- If you're thinking "search and view" - Use `sourcegraph` THEN `view` (two calls)
- If you're thinking "list and view" - Use `ls` THEN `view` (two calls)
- NEVER combine operations into one tool name

### ERROR RECOVERY:
When you see "Tool not found":
1. STOP trying the same tool name
2. LOOK at the tool name you used
3. IDENTIFY the operations you're trying to combine
4. SPLIT into separate tool calls
5. USE the exact tool names from the list

Remember: Each tool does ONE thing. To do multiple things, use multiple tools in sequence.