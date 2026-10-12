package hooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/AriaPutra01/agyless/internal/config"
)

type toolCallRequest struct {
	ToolCall struct {
		Name string                 `json:"name"`
		Args map[string]interface{} `json:"args"`
	} `json:"toolCall"`
}

type toolCallResponse struct {
	Decision  string                 `json:"decision"`
	AllowTool bool                   `json:"allowTool"`
	Overwrite map[string]interface{} `json:"overwrite"`
}

// RunRtkHook handles PreToolUse hook for Antigravity's run_command.
func RunRtkHook() int {
	input, err := io.ReadAll(os.Stdin)
	if err != nil || len(input) == 0 {
		return 0
	}

	var req toolCallRequest
	if err := json.Unmarshal(input, &req); err != nil {
		return 0
	}

	if req.ToolCall.Name != "run_command" || req.ToolCall.Args == nil {
		return 0
	}

	cmdLine, ok := req.ToolCall.Args["CommandLine"].(string)
	if !ok || strings.TrimSpace(cmdLine) == "" {
		return 0
	}

	rtkPath := config.FindBinary("rtk")
	if rtkPath == "" {
		return 0
	}

	// Call rtk rewrite <cmdLine>
	cmd := exec.Command(rtkPath, "rewrite", cmdLine)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = io.Discard
	_ = cmd.Run()

	newCmd := strings.TrimSpace(stdout.String())
	if newCmd == "" || newCmd == cmdLine {
		return 0
	}

	req.ToolCall.Args["CommandLine"] = newCmd

	resp := toolCallResponse{
		Decision:  "allow",
		AllowTool: true,
		Overwrite: req.ToolCall.Args,
	}

	out, err := json.Marshal(resp)
	if err == nil {
		fmt.Println(string(out))
	}
	return 0
}
