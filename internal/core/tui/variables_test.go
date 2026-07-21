package tui

import (
	"reflect"
	"testing"
)

func TestHasVariables(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{"docker ps", false},
		{"docker logs -f {{container}}", true},
		{"docker exec -it {{container}} {{shell:sh}}", true},
		{"no variables here", false},
		{"lone {{ without close", false},
		{"lose }} without open", false},
		{"{{name:default}}", true},
		{"{{name}}", true},
		{"{{}}", true},
	}
	for _, tt := range tests {
		if got := HasVariables(tt.cmd); got != tt.want {
			t.Errorf("HasVariables(%q) = %v, want %v", tt.cmd, got, tt.want)
		}
	}
}

func TestParseVars(t *testing.T) {
	tests := []struct {
		cmd  string
		want []varInfo
	}{
		{"echo {{name}}", []varInfo{{name: "name", def: "", raw: "name"}}},
		{"echo {{name:default}}", []varInfo{{name: "name", def: "default", raw: "name:default"}}},
		{"{{a}} {{b:val}}", []varInfo{
			{name: "a", def: "", raw: "a"},
			{name: "b", def: "val", raw: "b:val"},
		}},
		{"no vars", nil},
		{"{{unclosed", nil},
		{"unopened}}", nil},
		{"{{ }}", []varInfo{{name: "", def: "", raw: " "}}},
		{"{{name:}}", []varInfo{{name: "name", def: "", raw: "name:"}}},
		{"{{ :default}}", []varInfo{{name: "", def: "default", raw: " :default"}}},
		{"{{a}}middle{{b}}", []varInfo{
			{name: "a", def: "", raw: "a"},
			{name: "b", def: "", raw: "b"},
		}},
		{"lone {{ here }} and {{another:val}} there", []varInfo{
			{name: "here", def: "", raw: " here "},
			{name: "another", def: "val", raw: "another:val"},
		}},
	}
	for _, tt := range tests {
		got := parseVars(tt.cmd)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("parseVars(%q) = %v, want %v", tt.cmd, got, tt.want)
		}
	}
}

func TestParseVarsTrimsName(t *testing.T) {
	vars := parseVars("{{  spaced  :default}}")
	if len(vars) != 1 {
		t.Fatalf("got %d vars", len(vars))
	}
	if vars[0].name != "spaced" {
		t.Fatalf("name = %q, want 'spaced'", vars[0].name)
	}
	if vars[0].def != "default" {
		t.Fatalf("def = %q, want 'default'", vars[0].def)
	}
}

func TestParseVarsWithColonInsideDefault(t *testing.T) {
	vars := parseVars("{{name:default:with:colons}}")
	if len(vars) != 1 {
		t.Fatalf("got %d vars", len(vars))
	}
	if vars[0].name != "name" {
		t.Fatalf("name = %q", vars[0].name)
	}
	if vars[0].def != "default:with:colons" {
		t.Fatalf("def = %q", vars[0].def)
	}
}
