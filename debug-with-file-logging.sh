#!/bin/bash

echo "Running OpenCode with file logging enabled"
echo "========================================="
echo ""

# Clear previous debug log
> ~/.opencode/debug.log

# Enable file logging
export OPENCODE_DEV_DEBUG=true
export OPENCODE_DEBUG=true

echo "Starting OpenCode..."
echo "Debug logs will be written to: ~/.opencode/debug.log"
echo "Tail the log file in another terminal with:"
echo "  tail -f ~/.opencode/debug.log"
echo ""
echo "Starting in 3 seconds..."
sleep 3

# Run opencode
./opencode -d