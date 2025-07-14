#!/bin/bash

echo "Testing OpenCode MCP startup debugging"
echo "======================================"
echo ""

# Set all debug flags
export OPENCODE_DEBUG=true
export OPENCODE_DEV_DEBUG=true

# Test 1: Run with MCP config
echo "Test 1: Running WITH MCP config..."
echo "Starting at: $(date)"
echo ""

# Run for just a few seconds to see startup logs
(
    ./opencode -d 2>&1 | while IFS= read -r line; do
        echo "$(date '+%H:%M:%S.%3N') $line"
    done
) &

PID=$!
echo "OpenCode PID: $PID"
echo ""
echo "Letting it run for 5 seconds to capture startup..."
sleep 5

# Kill it
echo ""
echo "Killing OpenCode..."
kill $PID 2>/dev/null
wait $PID 2>/dev/null

echo ""
echo "Test complete at: $(date)"