package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/matheuzgomes/decoreba/internal/core"
)

type execParams struct {
	ID        string `json:"id,omitempty"`
	Command   string `json:"command,omitempty"`
	Execute   bool   `json:"execute"`
	AllowSudo bool   `json:"allow_sudo"`
}

func handleExecute(s *core.Store, id json.RawMessage, args json.RawMessage) {
	var p execParams
	if err := json.Unmarshal(args, &p); err != nil {
		writeError(id, -32602, "Invalid params: "+err.Error())
		return
	}

	var cmd *core.Command
	var cmdStr string

	if p.ID != "" {
		var count int
		cmd, count = s.FindByPrefix(p.ID)
		if count == 0 {
			writeResult(id, map[string]interface{}{
				"content": textContent("No command found with that id."),
			})
			return
		}
		if count > 1 {
			writeResult(id, map[string]interface{}{
				"content": textContent(fmt.Sprintf("Ambiguous id %q (%d matches). Use more characters.", p.ID, count)),
			})
			return
		}
		cmdStr = cmd.Command
	} else if p.Command != "" {
		cmdStr = p.Command
	} else {
		writeError(id, -32602, "Either id or command is required")
		return
	}

	if entry, blocked := CheckBlocklist(cmdStr); blocked {
		if entry.Severity == 2 || !p.AllowSudo {
			writeResult(id, map[string]interface{}{
				"content": textContent(fmt.Sprintf("Blocked: %s\n\nPass \"allow_sudo\": true to override.", entry.Label)),
			})
			return
		}
	}

	if cmd != nil && cmd.IsWorkflow() {
		writeResult(id, map[string]interface{}{
			"content": textContent("Workflows cannot be executed via MCP. Execute individual steps instead."),
		})
		return
	}

	if !p.Execute {
		preview := "Dry-run (pass \"execute\": true to run):\n"
		preview += cmdStr
		if cmd != nil {
			preview += fmt.Sprintf("\n\nFrom: [%s] %s (%s)", cmd.ID, cmd.Title, cmd.Context)
		}
		writeResult(id, map[string]interface{}{
			"content": textContent(preview),
		})
		return
	}

	// Bump usage if it's a stored command
	if cmd != nil {
		for i := range s.Commands {
			if s.Commands[i].ID == cmd.ID {
				s.Commands[i].UsageCount++
				break
			}
		}
	}

	var stdout, stderr bytes.Buffer
	c := exec.Command("sh", "-c", cmdStr)
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			writeError(id, -32603, "Execution failed: "+err.Error())
			return
		}
	}

	result := map[string]interface{}{
		"content": textContent(fmt.Sprintf("Exit code: %d\n\nStdout:\n%s\n\nStderr:\n%s",
			exitCode, stdout.String(), stderr.String())),
	}
	writeResult(id, result)
}
