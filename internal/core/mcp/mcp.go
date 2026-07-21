package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/store"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func textContent(s string) []ToolContent {
	return []ToolContent{{Type: "text", Text: s}}
}

func Serve() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	s, err := store.Load()
	if err != nil {
		writeError(nil, -32603, "Failed to load store: "+err.Error())
		return
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			writeError(nil, -32700, "Parse error: "+err.Error())
			continue
		}

		handle(s, req)
	}

	if err := scanner.Err(); err != nil {
		writeError(nil, -32603, "Read error: "+err.Error())
	}
}

func handle(s *core.Store, req JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		handleInitialize(req)
	case "notifications/initialized":
	case "tools/list":
		handleToolsList(req)
	case "tools/call":
		handleToolsCall(s, req)
	default:
		writeError(req.ID, -32601, "Method not found: "+req.Method)
	}
}

func handleInitialize(req JSONRPCRequest) {
	var params struct {
		ProtocolVersion string `json:"protocolVersion"`
		ClientInfo      struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"clientInfo"`
	}
	if req.Params != nil {
		json.Unmarshal(req.Params, &params)
	}

	result := map[string]interface{}{
		"protocolVersion": "2025-11-25",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]string{
			"name":    "decoreba",
			"version": "0.3.0",
		},
	}
	writeResult(req.ID, result)
}

func handleToolsList(req JSONRPCRequest) {
	tools := []ToolDefinition{
		{
			Name:        "decoreba_search",
			Description: "Search commands with fuzzy matching, typo tolerance, and recency ranking.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query":   map[string]string{"type": "string", "description": "Search query"},
					"context": map[string]string{"type": "string", "description": "Filter by context (optional)"},
					"limit":   map[string]interface{}{"type": "integer", "description": "Max results (default 10)"},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "decoreba_get",
			Description: "Get full command details by ID.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]string{"type": "string", "description": "Command ID or prefix"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "decoreba_list_contexts",
			Description: "List all contexts with command counts.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "decoreba_add",
			Description: "Add a new command. Returns preview before saving — pass confirm: true to commit.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"context": map[string]string{"type": "string", "description": "Context (e.g. docker, git)"},
					"title":   map[string]string{"type": "string", "description": "Short description"},
					"command": map[string]string{"type": "string", "description": "Shell command"},
					"tags":    map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}, "description": "Optional tags"},
					"notes":   map[string]string{"type": "string", "description": "Optional notes"},
					"confirm": map[string]interface{}{"type": "boolean", "description": "Pass true to save"},
				},
				"required": []string{"context", "title", "command"},
			},
		},
		{
			Name:        "decoreba_edit",
			Description: "Edit an existing command by ID. Returns preview of changes — pass confirm: true to commit.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id":      map[string]string{"type": "string", "description": "Command ID or prefix"},
					"context": map[string]string{"type": "string", "description": "New context (optional)"},
					"title":   map[string]string{"type": "string", "description": "New title (optional)"},
					"command": map[string]string{"type": "string", "description": "New command (optional)"},
					"tags":    map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}, "description": "New tags (optional)"},
					"notes":   map[string]string{"type": "string", "description": "New notes (optional)"},
					"confirm": map[string]interface{}{"type": "boolean", "description": "Pass true to save changes"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "decoreba_remove",
			Description: "Remove a command by ID. Shows preview first — pass confirm: true to permanently delete.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id":      map[string]string{"type": "string", "description": "Command ID or prefix"},
					"confirm": map[string]interface{}{"type": "boolean", "description": "Pass true to confirm deletion"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "decoreba_execute",
			Description: "Preview or execute a command by ID. Dry-run by default — pass execute: true to run. Blocked commands (rm -rf, mkfs, etc.) require allow_sudo: true.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id":         map[string]string{"type": "string", "description": "Command ID or prefix"},
					"execute":    map[string]interface{}{"type": "boolean", "description": "Pass true to actually execute"},
					"allow_sudo": map[string]interface{}{"type": "boolean", "description": "Override blocklist for sudo/safe patterns"},
				},
			},
		},
		{
			Name:        "decoreba_stats",
			Description: "Show vault statistics: total commands, per context, per tag, most used, recently used.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	writeResult(req.ID, map[string]interface{}{
		"tools": tools,
	})
}

func handleToolsCall(s *core.Store, req JSONRPCRequest) {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		writeError(req.ID, -32602, "Invalid params: "+err.Error())
		return
	}

	switch params.Name {
	case "decoreba_search":
		handleSearch(s, req.ID, params.Arguments)
	case "decoreba_get":
		handleGet(s, req.ID, params.Arguments)
	case "decoreba_list_contexts":
		handleListContexts(s, req.ID, params.Arguments)
	case "decoreba_add":
		handleAdd(s, req.ID, params.Arguments)
	case "decoreba_edit":
		handleEdit(s, req.ID, params.Arguments)
	case "decoreba_remove":
		handleRemove(s, req.ID, params.Arguments)
	case "decoreba_execute":
		handleExecute(s, req.ID, params.Arguments)
	case "decoreba_stats":
		handleStats(s, req.ID, params.Arguments)
	default:
		writeError(req.ID, -32602, "Unknown tool: "+params.Name)
	}
}

func writeError(id json.RawMessage, code int, msg string) {
	var idVal interface{}
	if id != nil {
		json.Unmarshal(id, &idVal)
	}
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      idVal,
		Error:   &MCPError{Code: code, Message: msg},
	}
	out, _ := json.Marshal(resp)
	fmt.Println(string(out))
}

func writeResult(id json.RawMessage, result interface{}) {
	var idVal interface{}
	if id != nil {
		json.Unmarshal(id, &idVal)
	}
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      idVal,
		Result:  result,
	}
	out, _ := json.Marshal(resp)
	fmt.Println(string(out))
}
