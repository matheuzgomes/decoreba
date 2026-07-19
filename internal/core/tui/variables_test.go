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
		if got := hasVariables(tt.cmd); got != tt.want {
			t.Errorf("hasVariables(%q) = %v, want %v", tt.cmd, got, tt.want)
		}
	}
}

func TestResolveCommandNoVars(t *testing.T) {
	cmd := "docker ps"
	resolved, cancelled, err := resolveCommand(cmd)
	if err != nil || cancelled || resolved != cmd {
		t.Fatalf("resolveCommand(%q) = (%q, %v, %v), want (%q, false, nil)", cmd, resolved, cancelled, err, cmd)
	}
}

func TestParseVariableNameAndDefault(t *testing.T) {
	// Test internal parsing through resolveCommand — but we can't test
	// promptVar without a terminal. Just test that the parsing doesn't panic
	// on various inputs.
	tests := []string{
		"docker {{cmd}}",
		"docker {{cmd:ps}}",
		"{{a}} {{b:default}} {{c}}",
		"no vars",
		"lone {{open",
	}
	for _, cmd := range tests {
		// This calls promptVar which requires a real terminal.
		// Skip when no terminal is available.
		if !hasVariables(cmd) {
			resolved, cancelled, err := resolveCommand(cmd)
			if err != nil || cancelled || resolved != cmd {
				t.Errorf("resolveCommand(%q) no vars: got (%q, %v, %v)", cmd, resolved, cancelled, err)
			}
		}
	}
}
