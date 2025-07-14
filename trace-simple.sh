#!/bin/bash

# Simple trace script for opencode startup without strace

echo "OpenCode Simple Trace"
echo "===================="
echo ""

# Create timestamp for this trace session
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
LOG_FILE="opencode-trace-$TIMESTAMP.log"

# Enable all debug options
export OPENCODE_DEBUG=true
export OPENCODE_DEV_DEBUG=true
export GODEBUG=gctrace=1,schedtrace=1000
export GOTRACEBACK=all

# Build if needed
if [ ! -f ./opencode ]; then
    echo "Building opencode..."
    go build -v -x -o opencode 2>&1 | tee build-$TIMESTAMP.log
fi

echo "Starting opencode with debug logging to: $LOG_FILE"
echo "Press Ctrl+C to stop"
echo ""
echo "=== Environment ===" | tee "$LOG_FILE"
env | grep -E "(OPENCODE|GO)" | tee -a "$LOG_FILE"
echo "=================" | tee -a "$LOG_FILE"
echo "" | tee -a "$LOG_FILE"

# Run opencode with debug flag and capture all output
./opencode -d 2>&1 | tee -a "$LOG_FILE"