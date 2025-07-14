#!/bin/bash

# Test script to verify request display functionality

echo "Testing request display functionality..."
echo "1. Start OpenCode"
echo "2. Send a message to trigger a request"
echo "3. Check that request info appears below LSP Configuration"
echo "4. Switch models to verify the request info updates"
echo "5. Test error handling by canceling a request"

# Monitor the requests.log file
echo ""
echo "Monitoring requests.log for activity..."
tail -f ~/.opencode/requests.log &
TAIL_PID=$!

# Give user instructions
echo ""
echo "Press Ctrl+C to stop monitoring"
echo ""

# Wait for user to stop
trap "kill $TAIL_PID 2>/dev/null; exit" INT
wait