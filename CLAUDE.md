# OpenCode Development Guide

This is a fork of the OpenCode project maintained at https://github.com/AeyeOps/opencode.git

## Repository Information

- **Fork URL**: https://github.com/AeyeOps/opencode.git
- **Original Repository**: https://github.com/opencode-ai/opencode
- **Branch**: main

## Building OpenCode

To build the OpenCode binary:

```bash
go build -o opencode
```

This creates an executable binary named `opencode` in the current directory.

## Git Remote Configuration

The git remote is configured to point to the AeyeOps fork:

```
origin  https://github.com/AeyeOps/opencode.git (fetch)
origin  https://github.com/AeyeOps/opencode.git (push)
```

## Development Notes

- This is a Go project using Go 1.24.0
- The main entry point is `main.go`
- Configuration handling is in `internal/config/`
- LLM providers are implemented in `internal/llm/provider/`
- TUI components are in `internal/tui/`

## Current Status

The repository has uncommitted changes including XAI/Grok integration work and various documentation files.