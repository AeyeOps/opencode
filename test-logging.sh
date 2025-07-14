#!/bin/bash

echo "Testing OpenCode logging destinations"
echo "===================================="
echo ""

# Test 1: Standard debug mode (console only)
echo "Test 1: Standard debug mode (-d flag)"
echo "-------------------------------------"
echo "Running: ./opencode -d"
echo ""
timeout 2 ./opencode -d 2>&1 | head -20
echo ""

# Test 2: With OPENCODE_DEV_DEBUG (file logging)
echo "Test 2: Dev debug mode (OPENCODE_DEV_DEBUG=true)"
echo "-----------------------------------------------"
echo "This should write to ~/.opencode/debug.log"
echo ""

# Clear the debug log first
> ~/.opencode/debug.log

export OPENCODE_DEV_DEBUG=true
timeout 2 ./opencode -d 2>&1 | head -5

echo ""
echo "Contents of ~/.opencode/debug.log:"
echo "----------------------------------"
head -20 ~/.opencode/debug.log || echo "No debug.log found"

echo ""
echo "Summary:"
echo "--------"
echo "1. With just -d flag: Logs appear in console via the TUI"
echo "2. With OPENCODE_DEV_DEBUG=true: Logs go to ~/.opencode/debug.log"
echo "3. The logging you added will appear in:"
echo "   - Console when using -d flag (but mixed with TUI output)"
echo "   - ~/.opencode/debug.log when OPENCODE_DEV_DEBUG=true is set"