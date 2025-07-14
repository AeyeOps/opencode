#!/bin/bash

# Enable debug logging for opencode
export OPENCODE_DEBUG=true
export OPENCODE_DEV_DEBUG=true

# Build if needed
if [ ! -f ./opencode ]; then
    echo "Building opencode..."
    go build -o opencode
fi

echo "Starting opencode with full debug logging..."
echo "Debug logs will be written to:"
echo "- Console output (structured logs)"
echo "- ~/.opencode/logs/ directory (if configured)"
echo ""

# Run opencode with debug flag
./opencode -d