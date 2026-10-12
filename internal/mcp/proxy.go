package mcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

var allowedTools = map[string]bool{
	"ctx_batch_execute":   true,
	"ctx_execute":         true,
	"ctx_execute_file":    true,
	"ctx_index":           true,
	"ctx_search":          true,
	"ctx_fetch_and_index": true,
}

type framingMode int

const (
	framingNDJSON framingMode = iota
	framingContentLength
)

// RunProxy starts the child MCP process and filters tools/list and tools/call.
func RunProxy(argv []string) int {
	if len(argv) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no command provided to run-mcp")
		return 1
	}

	exe := argv[0]
	args := argv[1:]

	cmd := exec.Command(exe, args...)
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "StdinPipe error: %v\n", err)
		return 1
	}
	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "StdoutPipe error: %v\n", err)
		return 1
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Start error: %v\n", err)
		return 1
	}

	var outMu sync.Mutex
	writeToClient := func(data []byte, mode framingMode) error {
		outMu.Lock()
		defer outMu.Unlock()
		if mode == framingContentLength {
			if _, err := fmt.Fprintf(os.Stdout, "Content-Length: %d\r\n\r\n", len(data)); err != nil {
				return err
			}
			if _, err := os.Stdout.Write(data); err != nil {
				return err
			}
			return nil
		}
		if _, err := os.Stdout.Write(append(data, '\n')); err != nil {
			return err
		}
		return nil
	}

	// Track request IDs that called tools/list so we filter their responses
	listReqMu := sync.Mutex{}
	listReqIDs := map[string]bool{}

	// Goroutine 1: Client (Antigravity) -> Server (context-mode)
	go func() {
		defer stdin.Close()
		rd := bufio.NewReader(os.Stdin)
		for {
			msg, mode, err := readMessage(rd)
			if err != nil {
				return
			}

			var req struct {
				ID     json.RawMessage `json:"id"`
				Method string          `json:"method"`
				Params struct {
					Name string `json:"name"`
				} `json:"params"`
			}

			if json.Unmarshal(msg, &req) == nil {
				if req.Method == "tools/list" && len(req.ID) > 0 {
					listReqMu.Lock()
					listReqIDs[string(req.ID)] = true
					listReqMu.Unlock()
				} else if req.Method == "tools/call" && !allowedTools[req.Params.Name] {
					// Deny tool call not in allowlist
					errResp, _ := json.Marshal(map[string]interface{}{
						"jsonrpc": "2.0",
						"id":      req.ID,
						"error": map[string]interface{}{
							"code":    -32601,
							"message": fmt.Sprintf("tool %q is not available", req.Params.Name),
						},
					})
					_ = writeToClient(errResp, mode)
					continue
				}
			}

			// Forward to child process stdin
			if mode == framingContentLength {
				fmt.Fprintf(stdin, "Content-Length: %d\r\n\r\n", len(msg))
				stdin.Write(msg)
			} else {
				stdin.Write(append(msg, '\n'))
			}
		}
	}()

	// Loop 2: Server (context-mode) -> Client (Antigravity)
	rd := bufio.NewReader(stdout)
	for {
		msg, mode, err := readMessage(rd)
		if err != nil {
			break
		}

		var resp struct {
			ID     json.RawMessage `json:"id"`
			Result struct {
				Tools []map[string]interface{} `json:"tools"`
			} `json:"result"`
		}

		if json.Unmarshal(msg, &resp) == nil && len(resp.ID) > 0 {
			listReqMu.Lock()
			isListResp := listReqIDs[string(resp.ID)]
			if isListResp {
				delete(listReqIDs, string(resp.ID))
			}
			listReqMu.Unlock()

			if isListResp && resp.Result.Tools != nil {
				// Filter tools array
				var filteredTools []map[string]interface{}
				for _, tool := range resp.Result.Tools {
					if name, ok := tool["name"].(string); ok && allowedTools[name] {
						filteredTools = append(filteredTools, tool)
					}
				}

				var fullResp map[string]interface{}
				if json.Unmarshal(msg, &fullResp) == nil {
					if resMap, ok := fullResp["result"].(map[string]interface{}); ok {
						resMap["tools"] = filteredTools
						if modMsg, err := json.Marshal(fullResp); err == nil {
							msg = modMsg
						}
					}
				}
			}
		}

		if err := writeToClient(msg, mode); err != nil {
			break
		}
	}

	_ = cmd.Wait()
	if cmd.ProcessState != nil {
		return cmd.ProcessState.ExitCode()
	}
	return 0
}

func readMessage(r *bufio.Reader) ([]byte, framingMode, error) {
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			return nil, framingNDJSON, err
		}

		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}

		if strings.HasPrefix(strings.ToLower(string(trimmed)), "content-length:") {
			parts := strings.SplitN(string(trimmed), ":", 2)
			if len(parts) != 2 {
				return nil, framingContentLength, fmt.Errorf("invalid Content-Length header")
			}
			length, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil || length < 0 {
				return nil, framingContentLength, fmt.Errorf("invalid length: %v", err)
			}

			// Read headers until blank line
			for {
				hdr, err := r.ReadBytes('\n')
				if err != nil {
					return nil, framingContentLength, err
				}
				if len(bytes.TrimSpace(hdr)) == 0 {
					break
				}
			}

			buf := make([]byte, length)
			if _, err := io.ReadFull(r, buf); err != nil {
				return nil, framingContentLength, err
			}
			return buf, framingContentLength, nil
		}

		// NDJSON
		return bytes.TrimRight(line, "\r\n"), framingNDJSON, nil
	}
}
