package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed rules/GEMINI.md
var geminiRulesTemplate string

// InstallConfig writes hooks.json, mcp_config.json, settings.json, and GEMINI.md.
func InstallConfig() error {
	agylessPath := AgylessBin()
	codegraphPath := FindBinary("codegraph")
	if codegraphPath == "" {
		codegraphPath = "codegraph"
	}
	contextModePath := FindBinary("context-mode")
	if contextModePath == "" {
		contextModePath = "context-mode"
	}

	home := Home()
	geminiDir := filepath.Join(home, ".gemini")
	geminiConfigDir := filepath.Join(geminiDir, "config")
	geminiCliDir := filepath.Join(geminiDir, "antigravity-cli")

	for _, dir := range []string{geminiDir, geminiConfigDir, geminiCliDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}

	// 1. Write GEMINI.md
	geminiMdPath := filepath.Join(geminiDir, "GEMINI.md")
	if err := os.WriteFile(geminiMdPath, []byte(geminiRulesTemplate), 0644); err != nil {
		return fmt.Errorf("write %s: %w", geminiMdPath, err)
	}
	fmt.Printf("✓ Rules installed: %s\n", geminiMdPath)

	// 2. Write hooks.json
	hooksPath := filepath.Join(geminiConfigDir, "hooks.json")
	hooksData := map[string]interface{}{
		"rtk": map[string]interface{}{
			"PreToolUse": []interface{}{
				map[string]interface{}{
					"matcher": "",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": fmt.Sprintf("%s rtk-hook agy", agylessPath),
							"timeout": 10,
						},
					},
				},
			},
		},
		"agyless-codegraph-index": map[string]interface{}{
			"PostToolUse": []interface{}{
				map[string]interface{}{
					"matcher": "",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": fmt.Sprintf("%s agy-hook codegraph-index", agylessPath),
							"timeout": 120,
						},
					},
				},
			},
			"PreInvocation": []interface{}{
				map[string]interface{}{
					"type":    "command",
					"command": fmt.Sprintf("%s agy-hook codegraph-index", agylessPath),
					"timeout": 120,
				},
			},
		},
	}
	if err := writeJSONPretty(hooksPath, hooksData); err != nil {
		return fmt.Errorf("write %s: %w", hooksPath, err)
	}
	fmt.Printf("✓ Hooks installed: %s\n", hooksPath)

	// 3. Write mcp_config.json
	mcpData := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"codegraph": map[string]interface{}{
				"command": codegraphPath,
				"args":    []string{"serve", "--mcp"},
				"trust":   true,
			},
			"context-mode": map[string]interface{}{
				"command": agylessPath,
				"args": []string{
					"run-mcp",
					"--context-mode",
					contextModePath,
				},
				"trust": true,
			},
		},
	}

	for _, mcpPath := range []string{
		filepath.Join(geminiConfigDir, "mcp_config.json"),
		filepath.Join(geminiCliDir, "mcp_config.json"),
	} {
		if err := writeJSONPretty(mcpPath, mcpData); err != nil {
			return fmt.Errorf("write %s: %w", mcpPath, err)
		}
		fmt.Printf("✓ MCP config installed: %s\n", mcpPath)
	}

	// 4. Update settings.json (permissions & ui.compactToolOutput)
	for _, settingsPath := range []string{
		filepath.Join(geminiDir, "settings.json"),
		filepath.Join(geminiCliDir, "settings.json"),
	} {
		if err := updateSettingsFile(settingsPath); err != nil {
			return fmt.Errorf("update %s: %w", settingsPath, err)
		}
		fmt.Printf("✓ Settings updated: %s\n", settingsPath)
	}

	return nil
}

func updateSettingsFile(path string) error {
	cfg := map[string]interface{}{}
	if raw, err := os.ReadFile(path); err == nil && len(raw) > 0 {
		_ = json.Unmarshal(raw, &cfg)
	}

	// Update UI
	ui, ok := cfg["ui"].(map[string]interface{})
	if !ok {
		ui = map[string]interface{}{}
		cfg["ui"] = ui
	}
	ui["compactToolOutput"] = false

	// Update permissions
	perms, ok := cfg["permissions"].(map[string]interface{})
	if !ok {
		perms = map[string]interface{}{}
		cfg["permissions"] = perms
	}

	var allow []interface{}
	if existingAllow, ok := perms["allow"].([]interface{}); ok {
		allow = existingAllow
	}

	neededEntries := []string{
		"command(rtk)",
		"command(rtk )",
		"mcp(codegraph/*)",
		"mcp(context-mode/*)",
	}

	for _, needed := range neededEntries {
		found := false
		for _, item := range allow {
			if s, ok := item.(string); ok && s == needed {
				found = true
				break
			}
		}
		if !found {
			allow = append(allow, needed)
		}
	}
	perms["allow"] = allow

	return writeJSONPretty(path, cfg)
}

func writeJSONPretty(path string, v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0644)
}
