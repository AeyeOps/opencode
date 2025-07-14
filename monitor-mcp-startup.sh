#!/bin/bash

echo "Monitoring OpenCode MCP Startup"
echo "==============================="
echo ""

# Ensure logs directory exists
mkdir -p ./logs

# Clear previous logs
> ./logs/debug.log

# Set environment for file logging
export OPENCODE_DEV_DEBUG=true
export OPENCODE_DEBUG=true

echo "Starting OpenCode in background..."
./opencode -d > /dev/null 2>&1 &
OPENCODE_PID=$!

echo "OpenCode PID: $OPENCODE_PID"
echo ""
echo "Monitoring MCP initialization (press Ctrl+C to stop)..."
echo ""

# Monitor the log file for MCP-related messages
tail -f ./logs/debug.log | grep --line-buffered -E "(MCP|mcp|stdio|reflection|getTools|Initialize)" &
TAIL_PID=$!

# Clean up on exit
trap "kill $OPENCODE_PID $TAIL_PID 2>/dev/null; exit" INT TERM

# Wait for interrupt
wait