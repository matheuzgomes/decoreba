package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/matheuzgomes/decoreba/internal/core"
)

func writeResp(t *testing.T, id json.RawMessage, result interface{}) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	writeResult(id, result)

	w.Close()
	var buf strings.Builder
	// Read from r
	os.Stdout = old
	data := make([]byte, 4096)
	n, _ := r.Read(data)
	buf.Write(data[:n])
	return buf.String()
}

func writeErrResp(t *testing.T, id json.RawMessage, code int, msg string) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	writeError(id, code, msg)

	w.Close()
	os.Stdout = old
	var buf strings.Builder
	data := make([]byte, 4096)
	n, _ := r.Read(data)
	buf.Write(data[:n])
	return buf.String()
}

func TestWriteResultAndError(t *testing.T) {
	t.Run("result", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		writeResult(json.RawMessage(`1`), map[string]string{"ok": "yes"})

		w.Close()
		os.Stdout = old
		var buf strings.Builder
		data := make([]byte, 4096)
		n, _ := r.Read(data)
		buf.Write(data[:n])

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(buf.String()), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.ID.(float64) != 1 {
			t.Fatalf("id = %v, want 1", resp.ID)
		}
		if resp.Error != nil {
			t.Fatalf("unexpected error: %+v", resp.Error)
		}
	})

	t.Run("error", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		writeError(json.RawMessage(`"req1"`), -32601, "not found")

		w.Close()
		os.Stdout = old
		var buf strings.Builder
		data := make([]byte, 4096)
		n, _ := r.Read(data)
		buf.Write(data[:n])

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(buf.String()), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.ID.(string) != "req1" {
			t.Fatalf("id = %v, want req1", resp.ID)
		}
		if resp.Error == nil || resp.Error.Code != -32601 {
			t.Fatalf("error = %+v", resp.Error)
		}
	})
}

func TestHandleInitialize(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "initialize",
	}

	handle(nil, req) // uses writeResult internally
}

func TestHandleToolsList(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "tools/list",
	}

	handle(nil, req)
	// just verify no panic
}

func TestHandleUnknownMethod(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "unknown",
	}

	handle(nil, req)
}

// Helper to capture a handler's output into a JSONRPCResponse
func captureHandler(t *testing.T, s *core.Store, method string, args json.RawMessage) JSONRPCResponse {
	t.Helper()
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  method,
		Params:  args,
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	handle(s, req)

	w.Close()
	os.Stdout = old
	var buf strings.Builder
	data := make([]byte, 4096)
	n, _ := r.Read(data)
	buf.Write(data[:n])

	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(buf.String()), &resp); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, buf.String())
	}
	return resp
}

func TestHandleSearch(t *testing.T) {
	s := &core.Store{Commands: []core.Command{
		{ID: "1", Context: "git", Title: "Undo", Command: "git reset"},
		{ID: "2", Context: "docker", Title: "Prune", Command: "docker container prune"},
	}}

	t.Run("basic search", func(t *testing.T) {
		args, _ := json.Marshal(map[string]string{"query": "git"})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_search","arguments":`+string(args)+`}`))
		if resp.Error != nil {
			t.Fatalf("error: %+v", resp.Error)
		}
		result := resp.Result.(map[string]interface{})
		content := result["content"].([]interface{})[0].(map[string]interface{})
		text := content["text"].(string)
		if !strings.Contains(text, "git reset") {
			t.Fatalf("result missing command: %s", text)
		}
	})

	t.Run("no results", func(t *testing.T) {
		args, _ := json.Marshal(map[string]string{"query": "zzz"})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_search","arguments":`+string(args)+`}`))
		if resp.Error != nil {
			t.Fatalf("error: %+v", resp.Error)
		}
	})

	t.Run("empty query error", func(t *testing.T) {
		args, _ := json.Marshal(map[string]string{"query": ""})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_search","arguments":`+string(args)+`}`))
		if resp.Error == nil {
			t.Fatal("expected error for empty query")
		}
	})

	t.Run("context filter", func(t *testing.T) {
		args, _ := json.Marshal(map[string]interface{}{
			"query":   "prune",
			"context": "docker",
		})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_search","arguments":`+string(args)+`}`))
		if resp.Error != nil {
			t.Fatalf("error: %+v", resp.Error)
		}
	})
}

func TestHandleGet(t *testing.T) {
	s := &core.Store{Commands: []core.Command{
		{ID: "abc123", Context: "git", Title: "Undo", Command: "git reset"},
	}}

	t.Run("found", func(t *testing.T) {
		args, _ := json.Marshal(map[string]string{"id": "abc123"})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_get","arguments":`+string(args)+`}`))
		if resp.Error != nil {
			t.Fatalf("error: %+v", resp.Error)
		}
	})

	t.Run("not found", func(t *testing.T) {
		args, _ := json.Marshal(map[string]string{"id": "zzzz"})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_get","arguments":`+string(args)+`}`))
		res := resp.Result.(map[string]interface{})
		content := res["content"].([]interface{})[0].(map[string]interface{})
		if !strings.Contains(content["text"].(string), "No command found") {
			t.Fatalf("unexpected: %s", content["text"].(string))
		}
	})
}

func TestHandleListContexts(t *testing.T) {
	s := &core.Store{Commands: []core.Command{
		{ID: "1", Context: "git", Title: "x", Command: "y"},
		{ID: "2", Context: "git", Title: "z", Command: "w"},
		{ID: "3", Context: "docker", Title: "a", Command: "b"},
	}}

	resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_list_contexts","arguments":{}}`))
	if resp.Error != nil {
		t.Fatalf("error: %+v", resp.Error)
	}
}

func TestHandleAdd(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{}

	t.Run("preview without confirm", func(t *testing.T) {
		args, _ := json.Marshal(map[string]interface{}{
			"context": "git",
			"title":   "Undo",
			"command": "git reset",
		})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_add","arguments":`+string(args)+`}`))
		res := resp.Result.(map[string]interface{})
		content := res["content"].([]interface{})[0].(map[string]interface{})
		if !strings.Contains(content["text"].(string), "Will add") {
			t.Fatalf("preview missing: %s", content["text"].(string))
		}
		if !strings.Contains(content["text"].(string), "confirm") {
			t.Fatalf("preview should mention confirm: %s", content["text"].(string))
		}
		if len(s.Commands) != 0 {
			t.Fatal("command should not be saved without confirm")
		}
	})

	t.Run("save with confirm", func(t *testing.T) {
		args, _ := json.Marshal(map[string]interface{}{
			"context": "git",
			"title":   "Undo",
			"command": "git reset",
			"confirm": true,
		})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_add","arguments":`+string(args)+`}`))
		if resp.Error != nil {
			t.Fatalf("error: %+v", resp.Error)
		}
		if len(s.Commands) != 1 {
			t.Fatal("command should be saved")
		}
		if s.Commands[0].Context != "git" || s.Commands[0].Command != "git reset" {
			t.Fatalf("saved: %+v", s.Commands[0])
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		args, _ := json.Marshal(map[string]interface{}{
			"context": "",
			"title":   "x",
			"command": "y",
		})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_add","arguments":`+string(args)+`}`))
		if resp.Error == nil {
			t.Fatal("expected error for empty context")
		}
	})
}

func TestHandleEdit(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Commands: []core.Command{
		{ID: "abc123", Context: "git", Title: "Undo", Command: "git reset"},
	}}

	t.Run("preview without confirm", func(t *testing.T) {
		args, _ := json.Marshal(map[string]interface{}{
			"id":      "abc123",
			"command": "git reflog",
		})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_edit","arguments":`+string(args)+`}`))
		res := resp.Result.(map[string]interface{})
		content := res["content"].([]interface{})[0].(map[string]interface{})
		if !strings.Contains(content["text"].(string), "Will") && !strings.Contains(content["text"].(string), "Changes") {
			t.Fatalf("preview missing: %s", content["text"].(string))
		}
		if s.Commands[0].Command != "git reset" {
			t.Fatal("should not modify without confirm")
		}
	})

	t.Run("save with confirm", func(t *testing.T) {
		args, _ := json.Marshal(map[string]interface{}{
			"id":      "abc123",
			"command": "git reflog",
			"confirm": true,
		})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_edit","arguments":`+string(args)+`}`))
		if resp.Error != nil {
			t.Fatalf("error: %+v", resp.Error)
		}
		if s.Commands[0].Command != "git reflog" {
			t.Fatalf("command not updated: %+v", s.Commands[0])
		}
	})

	t.Run("no changes", func(t *testing.T) {
		args, _ := json.Marshal(map[string]interface{}{
			"id": "abc123",
		})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_edit","arguments":`+string(args)+`}`))
		res := resp.Result.(map[string]interface{})
		content := res["content"].([]interface{})[0].(map[string]interface{})
		if !strings.Contains(content["text"].(string), "No changes") {
			t.Fatalf("expected 'No changes': %s", content["text"].(string))
		}
	})

	t.Run("not found", func(t *testing.T) {
		args, _ := json.Marshal(map[string]interface{}{
			"id": "zzzz",
		})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_edit","arguments":`+string(args)+`}`))
		res := resp.Result.(map[string]interface{})
		content := res["content"].([]interface{})[0].(map[string]interface{})
		if !strings.Contains(content["text"].(string), "No command found") {
			t.Fatalf("unexpected: %s", content["text"].(string))
		}
	})
}

func TestHandleRemove(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Commands: []core.Command{
		{ID: "abc123", Context: "git", Title: "Undo", Command: "git reset"},
		{ID: "def456", Context: "docker", Title: "Prune", Command: "docker container prune"},
	}}

	t.Run("preview without confirm", func(t *testing.T) {
		args, _ := json.Marshal(map[string]interface{}{
			"id": "abc123",
		})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_remove","arguments":`+string(args)+`}`))
		res := resp.Result.(map[string]interface{})
		content := res["content"].([]interface{})[0].(map[string]interface{})
		if !strings.Contains(content["text"].(string), "Will delete") {
			t.Fatalf("preview missing: %s", content["text"].(string))
		}
		if len(s.Commands) != 2 {
			t.Fatal("should not delete without confirm")
		}
	})

	t.Run("delete with confirm", func(t *testing.T) {
		args, _ := json.Marshal(map[string]interface{}{
			"id":      "abc123",
			"confirm": true,
		})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_remove","arguments":`+string(args)+`}`))
		if resp.Error != nil {
			t.Fatalf("error: %+v", resp.Error)
		}
		if len(s.Commands) != 1 || s.Commands[0].ID != "def456" {
			t.Fatal("command not removed")
		}
	})

	t.Run("not found", func(t *testing.T) {
		args, _ := json.Marshal(map[string]interface{}{
			"id": "zzzz",
		})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_remove","arguments":`+string(args)+`}`))
		res := resp.Result.(map[string]interface{})
		content := res["content"].([]interface{})[0].(map[string]interface{})
		if !strings.Contains(content["text"].(string), "No command found") {
			t.Fatalf("unexpected: %s", content["text"].(string))
		}
	})
}

func TestHandleExecute(t *testing.T) {
	s := &core.Store{Commands: []core.Command{
		{ID: "abc123", Context: "git", Title: "Echo", Command: "echo hello"},
		{ID: "wf", Context: "git", Title: "Workflow", Command: "", Steps: []core.WorkflowStep{{Title: "x", Command: "echo hi"}}},
		{ID: "rmrf", Context: "danger", Title: "Bad", Command: "rm -rf /"},
	}}

	t.Run("dry run", func(t *testing.T) {
		args, _ := json.Marshal(map[string]interface{}{
			"id": "abc123",
		})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_execute","arguments":`+string(args)+`}`))
		res := resp.Result.(map[string]interface{})
		content := res["content"].([]interface{})[0].(map[string]interface{})
		if !strings.Contains(content["text"].(string), "Dry-run") {
			t.Fatalf("unexpected: %s", content["text"].(string))
		}
	})

	t.Run("workflow rejected", func(t *testing.T) {
		args, _ := json.Marshal(map[string]interface{}{
			"id": "wf",
		})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_execute","arguments":`+string(args)+`}`))
		res := resp.Result.(map[string]interface{})
		content := res["content"].([]interface{})[0].(map[string]interface{})
		if !strings.Contains(content["text"].(string), "Workflows cannot be executed") {
			t.Fatalf("unexpected: %s", content["text"].(string))
		}
	})

	t.Run("blocklist with no override", func(t *testing.T) {
		args, _ := json.Marshal(map[string]interface{}{
			"id": "rmrf",
		})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_execute","arguments":`+string(args)+`}`))
		res := resp.Result.(map[string]interface{})
		content := res["content"].([]interface{})[0].(map[string]interface{})
		if !strings.Contains(content["text"].(string), "Blocked") {
			t.Fatalf("unexpected: %s", content["text"].(string))
		}
	})

	t.Run("blocklist with severity 2 cannot be overridden", func(t *testing.T) {
		args, _ := json.Marshal(map[string]interface{}{
			"id":         "rmrf",
			"allow_sudo": true,
		})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_execute","arguments":`+string(args)+`}`))
		res := resp.Result.(map[string]interface{})
		content := res["content"].([]interface{})[0].(map[string]interface{})
		if !strings.Contains(content["text"].(string), "Blocked") {
			t.Fatalf("severity 2 block should not be overridable: %s", content["text"].(string))
		}
	})

	t.Run("direct command dry run", func(t *testing.T) {
		args, _ := json.Marshal(map[string]interface{}{
			"command": "echo hi",
		})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_execute","arguments":`+string(args)+`}`))
		res := resp.Result.(map[string]interface{})
		content := res["content"].([]interface{})[0].(map[string]interface{})
		if !strings.Contains(content["text"].(string), "Dry-run") {
			t.Fatalf("unexpected: %s", content["text"].(string))
		}
	})

	t.Run("no id or command", func(t *testing.T) {
		args, _ := json.Marshal(map[string]interface{}{})
		resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_execute","arguments":`+string(args)+`}`))
		if resp.Error == nil {
			t.Fatal("expected error for no id or command")
		}
	})
}

func TestHandleStats(t *testing.T) {
	s := &core.Store{Commands: []core.Command{
		{ID: "1", Context: "git", Title: "Undo", Command: "git reset", UsageCount: 5, Tags: []string{"undo"}},
		{ID: "2", Context: "docker", Title: "Prune", Command: "docker container prune", UsageCount: 2},
	}}

	resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"decoreba_stats","arguments":{}}`))
	if resp.Error != nil {
		t.Fatalf("error: %+v", resp.Error)
	}
}

func TestHandleToolsListCapture(t *testing.T) {
	resp := captureHandler(t, nil, "tools/list", nil)
	if resp.Error != nil {
		t.Fatalf("error: %+v", resp.Error)
	}
}

func TestHandleInitializeCapture(t *testing.T) {
	resp := captureHandler(t, nil, "initialize", nil)
	if resp.Error != nil {
		t.Fatalf("error: %+v", resp.Error)
	}
}

func TestUnknownTool(t *testing.T) {
	s := &core.Store{}
	args, _ := json.Marshal(map[string]interface{}{})
	resp := captureHandler(t, s, "tools/call", json.RawMessage(`{"name":"nonexistent","arguments":`+string(args)+`}`))
	if resp.Error == nil {
		t.Fatal("expected error for unknown tool")
	}
}

func TestTextContent(t *testing.T) {
	c := textContent("hello")
	if len(c) != 1 || c[0].Type != "text" || c[0].Text != "hello" {
		t.Fatalf("got %+v", c)
	}
}
