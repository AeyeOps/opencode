#!/bin/bash
# OpenCode Log Monitor - Multi-pane tmux dashboard for OPENCODE_DEV_DEBUG=true
# This script creates a tmux session with multiple panes to monitor OpenCode logs

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SESSION_NAME="opencode-monitor"
LOG_DIR="./logs"
DEBUG_LOG="$LOG_DIR/debug.log"
MESSAGES_DIR="$LOG_DIR/messages"

# Check if OPENCODE_DEV_DEBUG is set, if not set it
if [[ "$OPENCODE_DEV_DEBUG" != "true" ]]; then
    echo -e "${YELLOW}OPENCODE_DEV_DEBUG is not set. Setting it to 'true'...${NC}"
    export OPENCODE_DEV_DEBUG=true
    echo -e "${GREEN}✓ OPENCODE_DEV_DEBUG has been set to 'true'${NC}"
    echo ""
    echo -e "${BLUE}Note: This enables file-based logging for OpenCode.${NC}"
    echo -e "${BLUE}Logs will be written to ./logs/${NC}"
    echo ""
fi

# Ensure log directory exists
if [[ ! -d "$LOG_DIR" ]]; then
    echo -e "${YELLOW}Creating OpenCode log directory at $LOG_DIR${NC}"
    mkdir -p "$LOG_DIR"
fi

# Ensure messages directory exists
if [[ ! -d "$MESSAGES_DIR" ]]; then
    echo -e "${YELLOW}Creating messages directory at $MESSAGES_DIR${NC}"
    mkdir -p "$MESSAGES_DIR"
fi

# Create debug.log if it doesn't exist
if [[ ! -f "$DEBUG_LOG" ]]; then
    echo -e "${YELLOW}Creating debug.log file at $DEBUG_LOG${NC}"
    touch "$DEBUG_LOG"
    echo -e "${BLUE}Note: The log file is empty. It will populate when you run OpenCode.${NC}"
fi

# Check if tmux is installed
if ! command -v tmux &> /dev/null; then
    echo -e "${RED}Error: tmux is not installed${NC}"
    echo "Install tmux with: sudo apt-get install tmux"
    exit 1
fi

# Kill existing session if it exists
tmux kill-session -t "$SESSION_NAME" 2>/dev/null

echo -e "${BLUE}Creating OpenCode monitoring dashboard...${NC}"

# Create new tmux session
if ! tmux new-session -d -s "$SESSION_NAME" -n "OpenCode Logs"; then
    echo -e "${RED}Error: Failed to create tmux session${NC}"
    echo "Debug info:"
    echo "  TMUX env var: $TMUX"
    echo "  Session name: $SESSION_NAME"
    echo ""
    echo "If you're already in a tmux session, try:"
    echo "  1. Exit this tmux session first (Ctrl+b d)"
    echo "  2. Run the monitor script from outside tmux"
    echo ""
    echo "Or force creation with: tmux new-session -d -s $SESSION_NAME"
    exit 1
fi

# Configure tmux settings for this session
tmux set -t "$SESSION_NAME" mouse on
tmux set -t "$SESSION_NAME" history-limit 50000
tmux set -t "$SESSION_NAME" pane-border-status top
tmux set -t "$SESSION_NAME" pane-border-format " #[fg=white]#{pane_index}: #[fg=cyan]#{pane_title} "

# Create 4-pane layout
# First split horizontally
tmux split-window -h -t "$SESSION_NAME:0"
# Then split each half vertically
tmux split-window -v -t "$SESSION_NAME:0.0"
tmux split-window -v -t "$SESSION_NAME:0.2"

# Now we have 4 panes: 0, 1, 2, 3

# Send commands to each pane and press enter
# Pane 0 (top-left): Main debug log
tmux send-keys -t "$SESSION_NAME:0.0" "tail -f $DEBUG_LOG | grep --color=always -E '.*'"
tmux send-keys -t "$SESSION_NAME:0.0" C-m

# Pane 1 (bottom-left): MCP logs
tmux send-keys -t "$SESSION_NAME:0.1" "tail -f $DEBUG_LOG | grep --color=always -E 'MCP|mcp|stdio|reflection|getTools|Initialize'"
tmux send-keys -t "$SESSION_NAME:0.1" C-m

# Pane 2 (top-right): Session directories
tmux send-keys -t "$SESSION_NAME:0.2" "watch -n 1 'echo \"=== Sessions ===\"; ls -la $MESSAGES_DIR 2>/dev/null | tail -20'"
tmux send-keys -t "$SESSION_NAME:0.2" C-m

# Pane 3 (bottom-right): Errors and Warnings
tmux send-keys -t "$SESSION_NAME:0.3" "tail -f $DEBUG_LOG | grep --color=always -E 'ERROR|WARN|panic|failed'"
tmux send-keys -t "$SESSION_NAME:0.3" C-m

# Select first pane
tmux select-pane -t "$SESSION_NAME:0.0"

# Verify session exists
if ! tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
    echo -e "${RED}Error: Failed to create monitoring session${NC}"
    exit 1
fi

# Attach to the session
echo -e "${GREEN}OpenCode monitoring dashboard ready!${NC}"
echo ""
echo "Pane layout:"
echo "  [0: Debug Log    ] [2: Sessions     ]"
echo "  [1: MCP Logs     ] [3: Errors/Warns ]"
echo ""
echo "Commands:"
echo "  Ctrl+b arrow - Navigate panes"
echo "  Ctrl+b z     - Zoom current pane"
echo "  Ctrl+b d     - Detach from session"
echo ""

# Check if we're already in a tmux session
if [[ -n "$TMUX" ]]; then
    echo -e "${YELLOW}Warning: You're already in a tmux session.${NC}"
    echo -e "${YELLOW}The monitoring session has been created as '$SESSION_NAME'.${NC}"
    echo ""
    echo "To switch to it:"
    echo "  tmux switch-client -t $SESSION_NAME"
    echo ""
    echo "Or detach from current session first:"
    echo "  Ctrl+b d"
    echo "  tmux attach -t $SESSION_NAME"
else
    tmux attach-session -t "$SESSION_NAME"
fi