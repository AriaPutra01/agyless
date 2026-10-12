# agyless

Token saver & MCP/Hook orchestrator built specifically for Google Antigravity AI Coding Agent (`agy`).

Refactored, minimal, and standalone rewrite of the token optimization pipeline — stripped of legacy multi-agent overhead and customized with bug-free Antigravity hook wiring.

---

## ⚡ Quick Start

### Option 1: One-Line Installer (macOS / Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/AriaPutra01/agyless/main/scripts/install.sh | bash
```

### Option 2: Via Go

```bash
go install github.com/AriaPutra01/agyless/cmd/agyless@latest
agyless init
```

### Option 3: From Source

```bash
git clone https://github.com/AriaPutra01/agyless.git
cd agyless
go build -o ~/.local/bin/agyless ./cmd/agyless
agyless init
```

---

## 🛠️ What `agyless` Does

`agyless init` automatically configures and wires four token-saving capabilities into Antigravity:

1. **RTK Command Rewriting Hook (`agyless rtk-hook agy`)**
   - Intercepts `run_command` in `PreToolUse`.
   - Rewrites shell commands via `rtk rewrite <cmd>` (e.g., `git status` -> `rtk git status`, `ls -la` -> `rtk ls -la`).
   - Drastically compresses, formats, and truncates terminal output to save agent context tokens.

2. **CodeGraph Auto-Indexer Hook (`agyless agy-hook codegraph-index`)**
   - Intercepts Antigravity sessions (`PreInvocation` and `PostToolUse`).
   - Automatically detects project roots and initializes `.codegraph` AST knowledge graphs.
   - Provides instant symbol search, call paths, and blast radius without reading raw files.

3. **Bounded Context-Mode MCP Proxy (`agyless run-mcp --context-mode`)**
   - Acts as a transparent JSON-RPC stdio proxy between Antigravity and the `context-mode` MCP server.
   - Restricts tool exposure strictly to 6 essential sandbox tools:
     - `ctx_batch_execute`
     - `ctx_execute`
     - `ctx_execute_file`
     - `ctx_index`
     - `ctx_search`
     - `ctx_fetch_and_index`
   - Rejects unneeded/token-heavy tools at proxy level.

4. **Global Agent Instructions (`~/.gemini/GEMINI.md`)**
   - Injects battle-tested behavioral guidelines (Principles, Caveman terse response style, Ponytail build discipline, CodeGraph usage ladder, Context-mode sandbox rules).

---

## 📋 CLI Commands

| Command | Description |
|---|---|
| `agyless` / `agyless init` | Install external deps and wire Antigravity hooks, MCP, settings, and rules. |
| `agyless index [path]` | Manually build or refresh CodeGraph AST index (`.codegraph/`) in the current or target directory. |
| `agyless version` | Display current version. |
| `agyless rtk-hook agy` | Internal: PreToolUse hook for `run_command`. |
| `agyless agy-hook codegraph-index` | Internal: PreInvocation & PostToolUse hook for CodeGraph auto-init. |
| `agyless run-mcp --context-mode <bin>` | Internal: Bounded JSON-RPC stdio proxy for `context-mode`. |

---

## ⚙️ Configuration Files Touched

When running `agyless init`, the following files in `~/.gemini` are configured:
- `~/.gemini/config/hooks.json`: Registers `rtk` PreToolUse hook and `agyless-codegraph-index` hooks.
- `~/.gemini/config/mcp_config.json`: Registers `codegraph` and `context-mode` MCP servers.
- `~/.gemini/settings.json`: Adds permissions auto-allow for RTK and MCP tools.
- `~/.gemini/GEMINI.md`: Writes global system prompt rules.

---

## 📄 License

MIT
