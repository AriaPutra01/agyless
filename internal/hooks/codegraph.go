package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AriaPutra01/agyless/internal/config"
)

type workspaceInput struct {
	WorkspacePaths []string `json:"workspacePaths"`
	WorkspaceRoots []string `json:"workspace_roots"`
	WorkspaceRoot  string   `json:"workspaceRoot"`
	Cwd            string   `json:"cwd"`
}

// RunCodegraphIndexHook handles agy-hook codegraph-index.
func RunCodegraphIndexHook() int {
	input, _ := io.ReadAll(os.Stdin)
	dir := resolveProjectDirFromInput(input)
	if dir == "" {
		return 0
	}
	return ensureCodegraphIndex(dir)
}

// RunManualIndex handles 'agyless index [dir]'.
func RunManualIndex(dir string) int {
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current working directory: %v\n", err)
			return 1
		}
	}
	dir = findProjectRoot(dir)
	codegraphPath := config.FindBinary("codegraph")
	if codegraphPath == "" {
		fmt.Println("Error: codegraph binary not found in PATH or ~/.local/bin")
		return 1
	}

	indexPath := filepath.Join(dir, ".codegraph")
	action := "init"
	if _, err := os.Stat(indexPath); err == nil {
		action = "index"
	}

	fmt.Printf("Running codegraph %s in %s...\n", action, dir)
	var cmd *exec.Cmd
	if action == "init" {
		cmd = exec.Command(codegraphPath, "init")
	} else {
		cmd = exec.Command(codegraphPath, "index")
	}
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("codegraph %s failed: %v\n", action, err)
		return 1
	}
	fmt.Printf("✓ CodeGraph index ready at %s\n", dir)
	return 0
}

func ensureCodegraphIndex(dir string) int {
	indexPath := filepath.Join(dir, ".codegraph")
	if _, err := os.Stat(indexPath); err == nil {
		// Index already exists, no need to re-init
		return 0
	}

	codegraphPath := config.FindBinary("codegraph")
	if codegraphPath == "" {
		return 0
	}

	cmd := exec.Command(codegraphPath, "init")
	cmd.Dir = dir
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	_ = cmd.Run()
	return 0
}

func resolveProjectDirFromInput(input []byte) string {
	var req workspaceInput
	if len(input) > 0 {
		_ = json.Unmarshal(input, &req)
	}

	var candidates []string
	candidates = append(candidates, req.WorkspaceRoots...)
	candidates = append(candidates, req.WorkspacePaths...)
	if req.WorkspaceRoot != "" {
		candidates = append(candidates, req.WorkspaceRoot)
	}
	if req.Cwd != "" {
		candidates = append(candidates, req.Cwd)
	}
	if len(candidates) == 0 {
		if cwd, err := os.Getwd(); err == nil {
			candidates = append(candidates, cwd)
		}
	}

	for _, cand := range candidates {
		cand = strings.TrimSpace(cand)
		if cand != "" {
			return findProjectRoot(cand)
		}
	}
	return ""
}

func findProjectRoot(startDir string) string {
	cur := filepath.Clean(startDir)
	for {
		for _, marker := range []string{".git", ".codegraph", "go.mod", "package.json", "Cargo.toml"} {
			if _, err := os.Stat(filepath.Join(cur, marker)); err == nil {
				return cur
			}
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	return filepath.Clean(startDir)
}
