package history

import (
	"os"
	"path/filepath"
	"testing"
)

// Existing tests — must keep passing.

func TestParseBash(t *testing.T) {
	got := parse("echo old\n#1712345678\nkubectl get pods\n", "bash")
	if got != "kubectl get pods" {
		t.Fatalf("got %q", got)
	}
}

func TestParseZshExtended(t *testing.T) {
	got := parse(": 1:0;echo old\n: 2:0;kubectl get pods\n", "zsh")
	if got != "kubectl get pods" {
		t.Fatalf("got %q", got)
	}
}

func TestParseFish(t *testing.T) {
	got := parse(`- cmd: echo old
  when: 1
- cmd: "kubectl get pods"
  when: 2
`, "fish")
	if got != "kubectl get pods" {
		t.Fatalf("got %q", got)
	}
}

func TestLastUsesConfiguredHistoryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history")
	if err := os.WriteFile(path, []byte("echo old\nkubectl get pods\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHELL", "/bin/bash")
	t.Setenv("HISTFILE", path)
	got, err := Last()
	if err != nil {
		t.Fatal(err)
	}
	if got != "kubectl get pods" {
		t.Fatalf("got %q", got)
	}
}

// New tests — isSelfCall

func TestIsSelfCallExact(t *testing.T) {
	if !isSelfCall("decoreba add --last") {
		t.Fatal("should match exact")
	}
}

func TestIsSelfCallFullPath(t *testing.T) {
	if !isSelfCall("/home/user/go/bin/decoreba add --last") {
		t.Fatal("should match full path")
	}
	if !isSelfCall("~/go/bin/decoreba add --last") {
		t.Fatal("should match tilde path")
	}
}

func TestIsSelfCallRelative(t *testing.T) {
	if !isSelfCall("./decoreba add --last") {
		t.Fatal("should match relative path")
	}
}

func TestIsSelfCallNotMatch(t *testing.T) {
	if isSelfCall("kubectl get pods") {
		t.Fatal("should not match unrelated command")
	}
	if isSelfCall("decoreba search") {
		t.Fatal("should not match other decoreba commands")
	}
	if isSelfCall("echo decoreba add --last") {
		t.Fatal("should not match echo of command")
	}
	if isSelfCall("decoreba") {
		t.Fatal("should not match bare executable")
	}
	if isSelfCall("") {
		t.Fatal("should not match empty")
	}
}

// New tests — LastExcludingSelf

func TestLastExcludingSelfZshSkipsSelf(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history")
	if err := os.WriteFile(path, []byte(": 1:0;kubectl get pods\n: 2:0;decoreba add --last\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHELL", "/bin/zsh")
	t.Setenv("HISTFILE", path)

	got, err := LastExcludingSelf()
	if err != nil {
		t.Fatal(err)
	}
	if got != "kubectl get pods" {
		t.Fatalf("got %q, want kubectl get pods", got)
	}
}

func TestLastExcludingSelfZshFullPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history")
	if err := os.WriteFile(path, []byte(": 1:0;kubectl get pods\n: 2:0;~/go/bin/decoreba add --last\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHELL", "/bin/zsh")
	t.Setenv("HISTFILE", path)

	got, err := LastExcludingSelf()
	if err != nil {
		t.Fatal(err)
	}
	if got != "kubectl get pods" {
		t.Fatalf("got %q", got)
	}
}

func TestLastExcludingSelfZshMultiline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history")
	if err := os.WriteFile(path, []byte(": 1:0;kubectl get pods --all-namespaces\n-o wide\n: 2:0;decoreba add --last\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHELL", "/bin/zsh")
	t.Setenv("HISTFILE", path)

	got, err := LastExcludingSelf()
	if err != nil {
		t.Fatal(err)
	}
	want := "kubectl get pods --all-namespaces\n-o wide"
	if got != want {
		t.Fatalf("got %q, want multiline %q", got, want)
	}
}

func TestLastExcludingSelfMultipleSelfCalls(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history")
	if err := os.WriteFile(path, []byte(": 1:0;git stash\n: 2:0;decoreba add --last\n: 3:0;decoreba add --last\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHELL", "/bin/zsh")
	t.Setenv("HISTFILE", path)

	got, err := LastExcludingSelf()
	if err != nil {
		t.Fatal(err)
	}
	if got != "git stash" {
		t.Fatalf("got %q, want git stash", got)
	}
}

func TestLastExcludingSelfBash(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history")
	if err := os.WriteFile(path, []byte("#1\necho old\n#2\ndecoreba add --last\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHELL", "/bin/bash")
	t.Setenv("HISTFILE", path)

	got, err := LastExcludingSelf()
	if err != nil {
		t.Fatal(err)
	}
	if got != "echo old" {
		t.Fatalf("got %q, want echo old", got)
	}
}

func TestLastExcludingSelfFish(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history")
	if err := os.WriteFile(path, []byte("- cmd: kubectl get pods\n  when: 1\n- cmd: decoreba add --last\n  when: 2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHELL", "/bin/fish")
	t.Setenv("HISTFILE", path)

	got, err := LastExcludingSelf()
	if err != nil {
		t.Fatal(err)
	}
	if got != "kubectl get pods" {
		t.Fatalf("got %q, want kubectl get pods", got)
	}
}

func TestLastExcludingSelfNoSelfCall(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history")
	if err := os.WriteFile(path, []byte(": 1:0;echo one\n: 2:0;echo two\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHELL", "/bin/zsh")
	t.Setenv("HISTFILE", path)

	got, err := LastExcludingSelf()
	if err != nil {
		t.Fatal(err)
	}
	if got != "echo two" {
		t.Fatalf("got %q, want echo two", got)
	}
}

func TestLastExcludingSelfOnlySelfCall(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history")
	if err := os.WriteFile(path, []byte(": 1:0;decoreba add --last\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHELL", "/bin/zsh")
	t.Setenv("HISTFILE", path)

	_, err := LastExcludingSelf()
	if err == nil {
		t.Fatal("expected error when only self-call exists")
	}
}

func TestLastExcludingSelfEmptyHistory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history")
	if err := os.WriteFile(path, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHELL", "/bin/zsh")
	t.Setenv("HISTFILE", path)

	_, err := LastExcludingSelf()
	if err == nil {
		t.Fatal("expected error on empty history")
	}
}

func TestLastExcludingSelfClearOrCd(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history")
	if err := os.WriteFile(path, []byte(": 1:0;clear\n: 2:0;cd ..\n: 3:0;decoreba add --last\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHELL", "/bin/zsh")
	t.Setenv("HISTFILE", path)

	got, err := LastExcludingSelf()
	if err != nil {
		t.Fatal(err)
	}
	if got != "cd .." {
		t.Fatalf("got %q, want cd .. (should not filter cd/clear)", got)
	}
}

// New tests — parseAll

func TestParseAllZshMultiple(t *testing.T) {
	entries := parseAll(": 1:0;echo one\n: 2:0;echo two\n: 3:0;echo three\n", "zsh")
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}
	if entries[0].Command != "echo one" {
		t.Errorf("entry 0 = %q", entries[0].Command)
	}
	if entries[2].Command != "echo three" {
		t.Errorf("entry 2 = %q", entries[2].Command)
	}
}

func TestParseAllBashMultiple(t *testing.T) {
	entries := parseAll("#1\necho one\n#2\necho two\n", "bash")
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
}

func TestParseAllFishMultiple(t *testing.T) {
	entries := parseAll("- cmd: echo one\n  when: 1\n- cmd: echo two\n  when: 2\n", "fish")
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
}
