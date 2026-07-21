package cli

import (
	"bufio"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(fn func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	r.Close()
	return string(out)
}

func mockStdin(t *testing.T, input string) func() {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.WriteString(input); err != nil {
		t.Fatal(err)
	}
	w.Close()
	orig := reader
	reader = bufio.NewReader(r)
	return func() {
		reader = orig
		r.Close()
	}
}

// contains checks if s contains substr (package-level for tests).
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
