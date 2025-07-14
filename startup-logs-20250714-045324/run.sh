#!/bin/bash
echo "=== Environment Variables ==="
env | grep -E "(OPENCODE|DEBUG|VERBOSE|LOG)" | sort
echo ""
echo "=== Starting opencode ==="
exec ./opencode -d "$@"
