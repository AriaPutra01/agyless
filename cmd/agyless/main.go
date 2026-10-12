package main

import (
	"fmt"
	"os"

	"github.com/AriaPutra01/agyless/internal/config"
	"github.com/AriaPutra01/agyless/internal/hooks"
	"github.com/AriaPutra01/agyless/internal/mcp"
)

const version = "1.0.0"

func helpText() string {
	return `agyless — Token saver & MCP/Hook orchestrator for Google Antigravity

Usage:
  agyless               Install and wire Antigravity hooks, MCP servers, and rules
  agyless init          Install and wire Antigravity hooks, MCP servers, and rules
  agyless index [dir]   Initialize/build CodeGraph AST index (.codegraph/) in directory
  agyless version       Show version

Internal Agent Hooks:
  agyless rtk-hook agy              PreToolUse hook for Antigravity run_command
  agyless agy-hook codegraph-index  PreInvocation/PostToolUse hook for CodeGraph auto-init
  agyless run-mcp --context-mode    Bounded MCP JSON-RPC proxy for context-mode
`
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		runInit()
		return
	}

	switch args[0] {
	case "init", "install":
		runInit()
	case "run-mcp":
		proxyArgs := args[1:]
		if len(proxyArgs) > 0 && proxyArgs[0] == "--context-mode" {
			proxyArgs = proxyArgs[1:]
		}
		os.Exit(mcp.RunProxy(proxyArgs))
	case "rtk-hook":
		if len(args) >= 2 && args[1] == "agy" {
			os.Exit(hooks.RunRtkHook())
		}
		fmt.Println(helpText())
		os.Exit(1)
	case "agy-hook":
		if len(args) >= 2 && args[1] == "codegraph-index" {
			os.Exit(hooks.RunCodegraphIndexHook())
		}
		fmt.Println(helpText())
		os.Exit(1)
	case "index":
		target := ""
		if len(args) >= 2 {
			target = args[1]
		}
		os.Exit(hooks.RunManualIndex(target))
	case "version", "-v", "--version":
		fmt.Println("agyless", version)
	case "help", "-h", "--help":
		fmt.Println(helpText())
	default:
		fmt.Printf("Unknown command: %s\n\n", args[0])
		fmt.Println(helpText())
		os.Exit(1)
	}
}

func runInit() {
	fmt.Println("=== agyless (Antigravity Toolkit) Setup ===")
	if err := config.EnsureDeps(); err != nil {
		fmt.Printf("Error checking dependencies: %v\n", err)
	}

	if err := config.InstallConfig(); err != nil {
		fmt.Printf("Error configuring Antigravity: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nSetup complete! Antigravity is now wired with:")
	fmt.Println("  • RTK command rewriting hook (PreToolUse)")
	fmt.Println("  • CodeGraph auto-indexing hook (PreInvocation & PostToolUse)")
	fmt.Println("  • Context-Mode filtered MCP proxy")
	fmt.Println("  • Global rules & instructions in ~/.gemini/GEMINI.md")
}
