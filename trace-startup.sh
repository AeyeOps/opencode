#!/bin/bash

# Script to capture full startup trace logs for opencode

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}OpenCode Startup Trace Tool${NC}"
echo "============================"
echo ""

# Create logs directory
LOG_DIR="./startup-logs-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$LOG_DIR"

echo -e "${YELLOW}Log directory:${NC} $LOG_DIR"
echo ""

# Set environment variables for maximum debugging
export OPENCODE_DEBUG=true
export OPENCODE_DEV_DEBUG=true
export OPENCODE_LOG_LEVEL=debug

# Also set generic debug vars that might be used
export DEBUG=true
export VERBOSE=true

# Build if needed
if [ ! -f ./opencode ]; then
    echo -e "${YELLOW}Building opencode...${NC}"
    go build -o opencode 2>&1 | tee "$LOG_DIR/build.log"
    if [ $? -ne 0 ]; then
        echo -e "${RED}Build failed. Check $LOG_DIR/build.log${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}Starting opencode with full trace logging...${NC}"
echo ""

# Create a wrapper script to capture all output
cat > "$LOG_DIR/run.sh" << 'EOF'
#!/bin/bash
echo "=== Environment Variables ==="
env | grep -E "(OPENCODE|DEBUG|VERBOSE|LOG)" | sort
echo ""
echo "=== Starting opencode ==="
exec ./opencode -d "$@"
EOF
chmod +x "$LOG_DIR/run.sh"

# Run with strace for system calls (if available)
if command -v strace &> /dev/null; then
    echo -e "${YELLOW}Running with strace (system call trace)...${NC}"
    strace -o "$LOG_DIR/strace.log" -f -t -s 1024 \
        "$LOG_DIR/run.sh" 2>&1 | tee "$LOG_DIR/opencode.log" &
    OPENCODE_PID=$!
else
    echo -e "${YELLOW}Running without strace (not available)...${NC}"
    "$LOG_DIR/run.sh" 2>&1 | tee "$LOG_DIR/opencode.log" &
    OPENCODE_PID=$!
fi

echo -e "${GREEN}OpenCode started with PID: $OPENCODE_PID${NC}"
echo ""
echo "Capturing startup logs for 10 seconds..."
echo "Press Ctrl+C to stop earlier if startup is complete"
echo ""

# Capture for 10 seconds or until interrupted
sleep 10

# Kill the opencode process
echo -e "${YELLOW}Stopping opencode...${NC}"
kill $OPENCODE_PID 2>/dev/null
wait $OPENCODE_PID 2>/dev/null

# Also capture any ~/.opencode logs if they exist
if [ -d ~/.opencode/logs ]; then
    echo -e "${YELLOW}Copying ~/.opencode/logs...${NC}"
    cp -r ~/.opencode/logs "$LOG_DIR/opencode-home-logs"
fi

# Generate summary
echo ""
echo -e "${GREEN}Trace Summary${NC}"
echo "============="
echo -e "Main log: ${YELLOW}$LOG_DIR/opencode.log${NC}"
if [ -f "$LOG_DIR/strace.log" ]; then
    echo -e "System calls: ${YELLOW}$LOG_DIR/strace.log${NC}"
fi
if [ -d "$LOG_DIR/opencode-home-logs" ]; then
    echo -e "Home logs: ${YELLOW}$LOG_DIR/opencode-home-logs${NC}"
fi

# Show first few debug lines
echo ""
echo -e "${GREEN}First debug log entries:${NC}"
grep -i -E "(debug|trace|info|warn|error)" "$LOG_DIR/opencode.log" | head -20

echo ""
echo -e "${GREEN}Trace capture complete!${NC}"