# OpenCode Logging Guide

This guide explains how to enable and use comprehensive logging in OpenCode for debugging and monitoring purposes.

## Quick Start

### Enable Debug Logging (Console Output)
```bash
./opencode -d
# or
./opencode --debug
```

### Enable File-Based Logging
```bash
export OPENCODE_DEV_DEBUG=true
./opencode
```

### Maximum Logging (Both Console and File)
```bash
export OPENCODE_DEV_DEBUG=true
./opencode -d
```

## Logging Methods Explained

### 1. Command Line Debug Flag (`-d` or `--debug`)
- Sets log level from INFO to DEBUG
- Outputs detailed logs to the console
- Includes source file:line numbers for traceability
- Best for quick debugging sessions

### 2. Environment Variable (`OPENCODE_DEV_DEBUG=true`)
- Enables persistent file-based logging
- Creates comprehensive log files and session records
- Logs all API requests/responses
- Persists across multiple runs
- Best for long-term debugging and analysis

### 3. Combined Approach
Using both methods together provides:
- Real-time console output for immediate feedback
- Persistent file logs for later analysis
- Complete debugging information

## Log File Locations

When `OPENCODE_DEV_DEBUG=true` is enabled:

| Log Type | Location | Description |
|----------|----------|-------------|
| Main Debug Log | `~/.opencode/debug.log` | All debug messages, errors, and warnings |
| Session Messages | `~/.opencode/messages/[session-id]/` | Detailed per-session logs |
| Request Logs | `~/.opencode/messages/[session-id]/[seq]_request.json` | API request payloads |
| Response Logs | `~/.opencode/messages/[session-id]/[seq]_response.json` | API response data |
| Stream Logs | `~/.opencode/messages/[session-id]/[seq]_response_stream.log` | Streaming responses |
| Tool Results | `~/.opencode/messages/[session-id]/[seq]_tool_results.json` | Tool execution results |

## Viewing Logs

### Real-Time Log Monitoring
```bash
# Watch the main debug log
tail -f ~/.opencode/debug.log

# Watch with highlighting
tail -f ~/.opencode/debug.log | grep --color=auto -E "ERROR|WARN|INFO|DEBUG"

# Watch session directory changes
watch -n 1 'ls -la ~/.opencode/messages/*/'
```

### Analyzing Session Logs
```bash
# View all requests in a session
find ~/.opencode/messages/[session-id]/ -name "*_request.json" -exec cat {} \; | jq .

# View all responses
find ~/.opencode/messages/[session-id]/ -name "*_response.json" -exec cat {} \; | jq .

# Search for errors across all logs
grep -r "ERROR" ~/.opencode/debug.log
```

## Log Format

### Debug Log Format
```
2024-01-15 10:23:45 DEBUG Main message source=/path/to/file.go:123 key=value
```

Components:
- **Timestamp**: When the log was created
- **Level**: DEBUG, INFO, WARN, ERROR
- **Message**: Main log message
- **Source**: File path and line number (in debug mode)
- **Attributes**: Key-value pairs with additional context

### Session Log Structure
```
~/.opencode/messages/
└── abc12345/                    # First 8 chars of session ID
    ├── 1_request.json          # First request
    ├── 1_response_stream.log   # Streaming response chunks
    ├── 1_response.json         # Final response
    ├── 1_tool_results.json     # Tool execution results
    ├── 2_request.json          # Second request
    └── ...
```

## Common Use Cases

### 1. Debugging API Errors
```bash
# Enable full logging
export OPENCODE_DEV_DEBUG=true
./opencode -d

# In another terminal
tail -f ~/.opencode/debug.log | grep -E "ERROR|provider|timeout"
```

### 2. Analyzing Tool Execution
```bash
# Find all tool results
find ~/.opencode/messages/ -name "*_tool_results.json" -mtime -1 | xargs cat | jq .
```

### 3. Monitoring Performance
```bash
# Check response times
grep "response time" ~/.opencode/debug.log | tail -20
```

### 4. Debugging Provider Issues
```bash
# Check provider initialization
grep -i "provider" ~/.opencode/debug.log | grep -E "init|load|error"
```

## Log Levels

OpenCode uses standard log levels:

| Level | Usage | Visibility |
|-------|-------|------------|
| DEBUG | Detailed information for debugging | Only with `-d` flag |
| INFO | General informational messages | Always visible |
| WARN | Warning messages | Always visible |
| ERROR | Error messages | Always visible |

## Advanced Configuration

### Custom Log Directory
Currently, logs are stored in `~/.opencode/`. To change this, you would need to modify the data directory configuration.

### Log Rotation
Log files can grow large. Consider implementing log rotation:
```bash
# Simple log rotation script
mv ~/.opencode/debug.log ~/.opencode/debug.log.$(date +%Y%m%d)
touch ~/.opencode/debug.log
```

### Cleaning Old Session Logs
```bash
# Remove session logs older than 7 days
find ~/.opencode/messages/ -type d -mtime +7 -exec rm -rf {} \;
```

## Troubleshooting

### Logs Not Appearing
1. Ensure `OPENCODE_DEV_DEBUG=true` is exported (not just set)
2. Check permissions on `~/.opencode/` directory
3. Verify disk space is available

### Too Much Logging
- Use `grep` to filter specific components
- Disable file logging by unsetting `OPENCODE_DEV_DEBUG`
- Use only `-d` flag for console-only debugging

### Finding Specific Errors
```bash
# Search for timeout errors
grep -i timeout ~/.opencode/debug.log

# Find panics
grep -A 10 -B 5 "panic" ~/.opencode/debug.log

# Look for specific session
grep "session-id-here" ~/.opencode/debug.log
```

## Best Practices

1. **Development**: Always use `OPENCODE_DEV_DEBUG=true` during development
2. **Production**: Use `-d` flag only when actively debugging
3. **Log Cleanup**: Regularly clean old session logs to save disk space
4. **Sensitive Data**: Be aware that logs may contain API keys or sensitive prompts
5. **Performance**: Extensive logging can impact performance; disable when not needed

## Integration with Other Tools

### Using with `jq` for JSON Analysis
```bash
# Pretty-print all requests
find ~/.opencode/messages/ -name "*_request.json" | xargs cat | jq .

# Extract specific fields
cat ~/.opencode/messages/*/1_response.json | jq '.choices[0].message.content'
```

### Log Aggregation
For production environments, consider sending logs to:
- ELK Stack (Elasticsearch, Logstash, Kibana)
- Splunk
- CloudWatch Logs
- Datadog

## Summary

OpenCode's logging system provides comprehensive debugging capabilities through:
- Console output with the `-d` flag
- Persistent file logging with `OPENCODE_DEV_DEBUG=true`
- Detailed session tracking with request/response logs
- Source code location tracking for easy debugging

Enable the appropriate level of logging based on your needs, and use the various log files to troubleshoot issues, analyze performance, and understand the application's behavior.