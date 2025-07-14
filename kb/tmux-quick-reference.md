# Tmux Quick Reference for OpenCode Monitoring

## Tmux Basics

### Starting Tmux
```bash
tmux new -s session-name    # Create named session
tmux ls                      # List sessions
tmux a -t session-name       # Attach to session
tmux kill-session -t name    # Kill session
```

### Key Bindings (Default prefix: Ctrl+b)
- `Ctrl+b %` - Split pane vertically
- `Ctrl+b "` - Split pane horizontally
- `Ctrl+b arrow` - Move between panes
- `Ctrl+b z` - Toggle pane zoom
- `Ctrl+b d` - Detach from session
- `Ctrl+b c` - Create new window
- `Ctrl+b n/p` - Next/previous window

### Pane Management
```bash
# Split panes from command line
tmux split-window -h         # Horizontal split
tmux split-window -v         # Vertical split
tmux split-window -h -p 30   # 30% width horizontal split
tmux split-window -v -p 25   # 25% height vertical split

# Send commands to panes
tmux send-keys -t pane-id 'command' Enter
tmux select-pane -t 0        # Select pane 0
```

### Layout Management
```bash
# Predefined layouts
tmux select-layout even-horizontal
tmux select-layout even-vertical
tmux select-layout main-horizontal
tmux select-layout main-vertical
tmux select-layout tiled
```

## Advanced Tmux for Monitoring

### Creating Complex Layouts
```bash
# Create a 4-pane monitoring layout
tmux new-session -d -s monitor
tmux split-window -h -p 50
tmux select-pane -t 0
tmux split-window -v -p 50
tmux select-pane -t 2
tmux split-window -v -p 50
```

### Running Commands in Panes
```bash
# Send commands to specific panes
tmux send-keys -t monitor:0.0 'tail -f file1.log' Enter
tmux send-keys -t monitor:0.1 'tail -f file2.log' Enter
tmux send-keys -t monitor:0.2 'watch -n 1 "ls -la"' Enter
tmux send-keys -t monitor:0.3 'htop' Enter
```

### Window and Pane Titles
```bash
# Set pane title
printf '\033]2;%s\033\\' "Log Monitor"

# In tmux config
set -g pane-border-status top
set -g pane-border-format "#{pane_index}: #{pane_title}"
```

### Synchronize Panes
```bash
# Type in all panes at once
tmux setw synchronize-panes on
# Turn off
tmux setw synchronize-panes off
```

## Useful Tmux Options for Monitoring

### Mouse Support
```bash
# Enable mouse (tmux 2.1+)
tmux set -g mouse on
```

### Scrollback Buffer
```bash
# Increase scrollback buffer
tmux set -g history-limit 50000
```

### Status Bar Customization
```bash
# Show session name, window, pane, date and time
tmux set -g status-right '#S #I:#P %d %b %R'
```

## Tmux for OpenCode Log Monitoring

### Essential Commands for Log Monitoring
```bash
# Create monitoring session
tmux new-session -d -s opencode-monitor

# Split into 4 panes
tmux split-window -h -p 50
tmux select-pane -t 0
tmux split-window -v -p 66
tmux select-pane -t 0
tmux split-window -v -p 50

# Label panes
tmux select-pane -t 0 -T "Main Debug Log"
tmux select-pane -t 1 -T "Session Activity"
tmux select-pane -t 2 -T "Latest Session"
tmux select-pane -t 3 -T "Error Filter"

# Run monitoring commands
tmux send-keys -t 0 'tail -f ~/.opencode/debug.log' Enter
tmux send-keys -t 1 'watch -n 1 "ls -la ~/.opencode/messages/"' Enter
tmux send-keys -t 2 'watch -n 1 "find ~/.opencode/messages -name \"*_request.json\" -mmin -5 | head -10"' Enter
tmux send-keys -t 3 'tail -f ~/.opencode/debug.log | grep --color=auto -E "ERROR|WARN"' Enter
```

### Quick Tips
1. Use `Ctrl+b [` to enter copy mode for scrolling
2. Use `Ctrl+b q` to show pane numbers
3. Use `Ctrl+b :` to enter command mode
4. Use `Ctrl+b !` to break pane into new window
5. Use `Ctrl+b x` to kill current pane

### Saving and Restoring Sessions
Consider using tmux-resurrect or tmux-continuum plugins for persistent sessions across reboots.

## Common Monitoring Patterns

### Pattern 1: Main + Filtered Views
```bash
# Main log + multiple filtered views
tmux send-keys -t 0 'tail -f app.log' Enter
tmux send-keys -t 1 'tail -f app.log | grep ERROR' Enter
tmux send-keys -t 2 'tail -f app.log | grep INFO' Enter
```

### Pattern 2: Time-based Monitoring
```bash
# Show recent activity
tmux send-keys -t 0 'watch -n 1 "find logs/ -mmin -10 -type f"' Enter
```

### Pattern 3: Multi-file Tail
```bash
# Using multitail if available
tmux send-keys -t 0 'multitail -i file1.log -i file2.log' Enter
```