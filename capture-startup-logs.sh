#!/bin/bash

echo "Capturing OpenCode startup logs with MCP debugging"
echo "================================================="
echo ""

# Set environment for maximum debugging
export OPENCODE_DEBUG=true
export OPENCODE_DEV_DEBUG=true

# Create a unique log file for this session
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
LOG_FILE="opencode-startup-$TIMESTAMP.log"

echo "Configuration file:"
cat ~/.config/opencode/.opencode.json | grep -A 20 mcpServers || echo "No MCP config found"
echo ""

echo "Starting OpenCode with debug logging..."
echo "Logs will be saved to: $LOG_FILE"
echo "Also check: ~/.opencode/debug.log"
echo ""
echo "Press Ctrl+C after a few seconds to stop..."
echo ""

# Run opencode and capture ALL output (stdout and stderr)
./opencode -d > "$LOG_FILE" 2>&1 &
PID=$!

echo "OpenCode started with PID: $PID"
echo ""

# Wait for user to interrupt or 10 seconds
trap "echo 'Stopping...'; kill $PID 2>/dev/null; exit" INT
sleep 10

# If we get here, kill it anyway
echo "Stopping OpenCode after 10 seconds..."
kill $PID 2>/dev/null

echo ""
echo "Startup log captured in: $LOG_FILE"
echo ""
echo "Searching for MCP-related messages:"
grep -i "mcp\|stdio\|reflection" "$LOG_FILE" || echo "No MCP messages found"

echo ""
echo "First 50 lines of log:"
head -50 "$LOG_FILE"