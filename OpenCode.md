# OpenCode Memory File

This file serves as a memory aid for the OpenCode AI assistant, containing key project information.

## Project Overview

OpenCode is a Go-based CLI application that brings AI assistance to your terminal. It provides a TUI for interacting with various AI models to help with coding tasks, debugging, and more.

Key features include:
- Interactive TUI built with Bubble Tea
- Support for multiple AI providers (OpenAI, Anthropic, Google Gemini, etc.)
- Session management and persistent storage using SQLite
- Tool integration for file operations, shell commands, and more
- LSP integration for code intelligence
- Custom commands and MCP support for extensibility

## Current Working Directory

/opt/aeo/opencode/

## Git Status

```
On branch main
Your branch is up to date with 'origin/main'.

Changes to be committed:
  (use "git restore --staged <file>..." to unstage)
	modified:   internal/config/config.go
	modified:   internal/llm/agent/agent.go
	modified:   internal/llm/models/xai.go
	new file:   internal/llm/models/xai.go.bak
	modified:   internal/llm/prompt/coder.go
	modified:   internal/llm/provider/openai.go
	modified:   internal/llm/provider/provider.go
	new file:   internal/llm/provider/xai.go
	modified:   internal/tui/components/dialog/models.go
	modified:   internal/tui/tui.go
	new file:   kb/code-structure.md
	new file:   kb/model-selection-flow.md
	new file:   kb/opencode-architecture.md
	new file:   kb/provider-visibility-issue.md
	new file:   kb/troubleshooting-guide.md

Changes not staged for commit:
  (use "git add/rm <file>..." to update what will be committed)
  (use "git restore <file>..." to discard changes in working directory)
	deleted:    .opencode.json
	modified:   cmd/root.go
	modified:   internal/config/config.go
	modified:   internal/llm/agent/agent.go
	modified:   internal/llm/models/models.go
	modified:   internal/llm/models/xai.go
	modified:   internal/llm/prompt/coder.go
	modified:   internal/llm/provider/anthropic.go
	modified:   internal/llm/provider/gemini.go
	modified:   internal/llm/provider/openai.go
	modified:   internal/llm/provider/provider.go
	modified:   internal/llm/provider/xai.go
	modified:   internal/llm/tools/sourcegraph.go
	modified:   internal/tui/components/chat/chat.go
	modified:   internal/tui/components/chat/list.go
	modified:   internal/tui/components/chat/message.go
	modified:   internal/tui/components/dialog/custom_commands.go
	modified:   internal/tui/components/dialog/models.go
	modified:   internal/tui/components/dialog/permission.go
	modified:   internal/tui/page/chat.go
	modified:   internal/tui/tui.go
	deleted:    kb/code-structure.md
	deleted:    kb/model-selection-flow.md
	deleted:    kb/opencode-architecture.md
	deleted:    kb/provider-visibility-issue.md
	deleted:    kb/troubleshooting-guide.md

Untracked files:
  (use "git add <file>..." to include in what will be committed)
	.claude/
	OpenCode.md
	docs-for-grok4-to-review/
	internal/llm/models/xai2.go
	internal/llm/provider/xai2.go
	internal/request/
	kb/xai-grok-4-prompt-recommendations.md
	repomix-output.xml
	repomix.config.json
	test_request_display.sh
```

## Recent Commits

f0571f5 - Aldehir Rojas, 13 days ago : fix(tool/grep): always show file names with rg (#271)
1f6eef4 - Gedy Palomino, 13 days ago : fix(mcp): ensure required field if nil (#278)
4427df5 - Tai Groot, 2 weeks ago : fixup early return for ollama (#266)
b9bedba - Bryan Vaz, 3 weeks ago : feat: add github copilot provider (#230)
73729ef - Kujtim Hoxha, 3 weeks ago : small readme update.

## Directory Structure

- /opt/aeo/opencode/
  - opt/
    - aeo/
      - opencode/
        - LICENSE
        - OpenCode.md
        - README.md
        - cmd/
          - root.go
          - schema/
            - README.md
            - main.go
        - docs-for-grok4-to-review/
          - code-structure.md
          - dynamic-tool-list-update.md
          - grok4-critical-issues.md
          - grok4-prompt-augmentation.md
          - grok4-system-prompt-template.md
          - model-selection-flow.md
          - next-improvements.md
          - opencode-architecture.md
          - provider-visibility-issue.md
          - steve-thoughts.md
          - timeout-error-analysis.md
          - troubleshooting-guide.md
          - xai2-integration-guide.md
        - go.mod
        - go.sum
        - install
        - internal/
          - app/
            - app.go
            - lsp.go
          - completions/
            - files-folders.go
          - config/
            - config.go
            - init.go
          - db/
            - connect.go
            - db.go
            - embed.go
            - files.sql.go
            - messages.sql.go
            - migrations/
              - 20250424200609_initial.sql
              - 20250515105448_add_summary_message_id.sql
            - models.go
            - querier.go
            - sessions.sql.go
            - sql/
              - files.sql
              - messages.sql
              - sessions.sql
          - diff/
            - diff.go
            - patch.go
          - fileutil/
            - fileutil.go
          - format/
            - format.go
            - spinner.go
          - history/
            - file.go
          - llm/
            - agent/
              - agent-tool.go
              - agent.go
              - mcp-tools.go
              - tools.go
            - models/
              - anthropic.go
              - azure.go
              - copilot.go
              - gemini.go
              - groq.go
              - local.go
              - models.go
              - openai.go
              - openrouter.go
              - vertexai.go
              - xai.go
              - xai.go.bak
              - xai2.go
            - prompt/
              - coder.go
              - prompt.go
              - prompt_test.go
              - summarizer.go
              - task.go
              - title.go
            - provider/
              - anthropic.go
              - azure.go
              - bedrock.go
              - copilot.go
              - gemini.go
              - openai.go
              - provider.go
              - vertexai.go
              - xai.go
              - xai2.go
            - tools/
              - bash.go
              - diagnostics.go
              - edit.go
              - fetch.go
              - file.go
              - glob.go
              - grep.go
              - ls.go
              - ls_test.go
              - patch.go
              - shell/
                - shell.go
              - sourcegraph.go
              - tools.go
              - view.go
              - write.go
          - logging/
            - logger.go
            - message.go
            - writer.go
          - lsp/
            - client.go
            - handlers.go
            - language.go
            - methods.go
            - protocol/
              - LICENSE
              - interface.go
              - pattern_interfaces.go
              - tables.go
              - tsdocument-changes.go
              - tsjson.go
              - tsprotocol.go
              - uri.go
            - protocol.go
            - transport.go
            - util/
              - edit.go
            - watcher/
              - watcher.go
          - message/
            - attachment.go
            - content.go
            - message.go
          - permission/
            - permission.go
          - pubsub/
            - broker.go
            - events.go
          - request/
            - tracker.go
          - session/
            - session.go
          - tui/
            - components/
              - chat/
                - chat.go
                - editor.go
                - list.go
                - message.go
                - sidebar.go
              - core/
                - status.go
              - dialog/
                - arguments.go
                - commands.go
                - complete.go
                - custom_commands.go
                - custom_commands_test.go
                - filepicker.go
                - help.go
                - init.go
                - models.go
                - permission.go
                - quit.go
                - session.go
                - theme.go
              - logs/
                - details.go
                - table.go
              - util/
                - simple-list.go
            - image/
              - images.go
            - layout/
              - container.go
              - layout.go
              - overlay.go
              - split.go
            - page/
              - chat.go
              - logs.go
              - page.go
            - styles/
              - background.go
              - icons.go
              - markdown.go
              - styles.go
            - theme/
              - catppuccin.go
              - dracula.go
              - flexoki.go
              - gruvbox.go
              - manager.go
              - monokai.go
              - onedark.go
              - opencode.go
              - theme.go
              - theme_test.go
              - tokyonight.go
              - tron.go
            - tui.go
            - util/
              - util.go
          - version/
            - version.go
        - kb/
          - xai-grok-4-prompt-recommendations.md
        - main.go
        - opencode
        - opencode-schema.json
        - repomix-output.xml
        - repomix.config.json
        - repomix.log
        - scripts/
          - check_hidden_chars.sh
          - release
          - snapshot
        - sqlc.yaml
        - test_request_display.sh

This memory file was rebuilt by the AI assistant.

