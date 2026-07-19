package tui

import (
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
	}
	for _, tt := range tests {
		if got := HasVariables(tt.cmd); got != tt.want {
			t.Errorf("HasVariables(%q) = %v, want %v", tt.cmd, got, tt.want)
		}
	}
}
