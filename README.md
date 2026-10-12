# tokless (Antigravity Edition)

Clean, lightweight, standalone toolkit focused exclusively on Google Antigravity AI Coding Agent (`agy`).

## Features
- **RTK Hook (`rtk-hook agy`)**: Intercepts `run_command` in `PreToolUse` to compress/truncate large CLI output using `rtk`.
- **CodeGraph Hook (`agy-hook codegraph-index`)**: Automatically initializes `.codegraph` AST index on session start (`PreInvocation`) and after tools (`PostToolUse`).
- **Context-Mode Bounded Proxy (`run-mcp --context-mode`)**: Restricts `context-mode` MCP tools strictly to the 6 essential sandbox tools (`ctx_batch_execute`, `ctx_execute`, `ctx_execute_file`, `ctx_index`, `ctx_search`, `ctx_fetch_and_index`).
- **Auto Configuration**: Automatically configures `~/.gemini/config/mcp_config.json`, `~/.gemini/config/hooks.json`, `~/.gemini/settings.json`, and `~/.gemini/GEMINI.md`.

## Build & Install

```bash
# Build binary
go build -o tokless ./cmd/tokless

# Move to PATH
mv tokless ~/.local/bin/tokless

# Run setup
tokless init
```

## Commands

- `tokless init`: Wire Antigravity configuration, hooks, and rules.
- `tokless index [path]`: Manually build or refresh CodeGraph index in directory.
- `tokless version`: Show version.
