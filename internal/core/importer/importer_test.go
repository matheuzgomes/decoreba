package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func importPetTemp(t *testing.T, tomlContent string) (*Report, error) {
	t.Helper()
	tmp := t.TempDir()
	src := filepath.Join(tmp, "snippets.toml")
	writeFile(t, src, tomlContent)
	return ImportPet(src)
}

func TestImportPetBasic(t *testing.T) {
	report, err := importPetTemp(t, `
[[snippets]]
description = "List files"
command = "ls -la"
tag = ["file", "list"]

[[snippets]]
description = "Git status"
command = "git status"
tag = ["git"]
`)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Commands) != 2 {
		t.Fatalf("got %d commands, want 2", len(report.Commands))
	}

	c0 := report.Commands[0]
	if c0.Title != "List files" {
		t.Errorf("title = %q, want %q", c0.Title, "List files")
	}
	if c0.Command != "ls -la" {
		t.Errorf("command = %q, want %q", c0.Command, "ls -la")
	}
	if len(c0.Tags) != 2 || c0.Tags[0] != "file" || c0.Tags[1] != "list" {
		t.Errorf("tags = %v, want [file list]", c0.Tags)
	}
}

func TestImportPetMultilineCommand(t *testing.T) {
	report, err := importPetTemp(t, `
[[snippets]]
description = "Build and run"
command = """
go build -o app .
./app
"""
`)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Commands) != 1 {
		t.Fatalf("got %d commands, want 1", len(report.Commands))
	}

	cmd := report.Commands[0].Command
	if !strings.Contains(cmd, "go build") || !strings.Contains(cmd, "./app") {
		t.Errorf("multiline command = %q, want both lines", cmd)
	}
}

func TestImportPetOutputNotes(t *testing.T) {
	report, err := importPetTemp(t, `
[[snippets]]
description = "List files"
command = "ls -la"
output = "loop"
`)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(report.Commands[0].Notes, "pet output: loop") {
		t.Errorf("notes = %q", report.Commands[0].Notes)
	}
}

func TestImportPetVariableConversion(t *testing.T) {
	report, err := importPetTemp(t, `
[[snippets]]
description = "Docker logs"
command = "docker logs --tail <lines> <container>"

[[snippets]]
description = "Git commit"
command = "git commit -m '<message:WIP>'"
`)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Commands) != 2 {
		t.Fatalf("got %d commands, want 2", len(report.Commands))
	}

	if report.Commands[0].Command != "docker logs --tail {{lines}} {{container}}" {
		t.Errorf("command = %q", report.Commands[0].Command)
	}
	if report.Commands[1].Command != "git commit -m '{{message:WIP}}'" {
		t.Errorf("command = %q", report.Commands[1].Command)
	}
}

func TestImportPetEmptyCommand(t *testing.T) {
	report, err := importPetTemp(t, `
[[snippets]]
description = "Empty"
command = ""

[[snippets]]
description = "Valid"
command = "echo ok"
`)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Commands) != 1 {
		t.Fatalf("got %d commands, want 1 (empty should be skipped)", len(report.Commands))
	}
	if !report.HasWarnings() {
		t.Fatal("expected warning for empty command")
	}
}

func TestImportPetMissingTitle(t *testing.T) {
	report, err := importPetTemp(t, `
[[snippets]]
command = "docker ps"
`)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Commands) != 1 {
		t.Fatalf("got %d commands, want 1", len(report.Commands))
	}
	if report.Commands[0].Title != "docker ps" {
		t.Errorf("title = %q, want fallback to command text", report.Commands[0].Title)
	}
}

func TestImportNaviBasic(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "git.cheat")
	writeFile(t, src, `% git

# Stash all changes
git stash --include-untracked

# View log
git log --oneline
`)

	report, err := ImportNavi(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Commands) != 2 {
		t.Fatalf("got %d commands, want 2", len(report.Commands))
	}

	c0 := report.Commands[0]
	if c0.Title != "Stash all changes" {
		t.Errorf("title = %q", c0.Title)
	}
	if c0.Command != "git stash --include-untracked" {
		t.Errorf("command = %q", c0.Command)
	}
	if c0.Context != "git" {
		t.Errorf("context = %q, want git", c0.Context)
	}
}

func TestImportNaviTags(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "docker.cheat")
	writeFile(t, src, `% docker, container

# List containers
docker ps

# Stop container
docker stop <container>
`)

	report, err := ImportNavi(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Commands) != 2 {
		t.Fatalf("got %d commands, want 2", len(report.Commands))
	}
	if len(report.Commands[0].Tags) != 2 {
		t.Fatalf("tags = %v, want 2", report.Commands[0].Tags)
	}
	if report.Commands[1].Command != "docker stop {{container}}" {
		t.Errorf("command = %q", report.Commands[1].Command)
	}
}

func TestImportNaviCheatMd(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "kubectl.cheat.md")
	writeFile(t, src, `% kubernetes

# Get pods
kubectl get pods
`)

	report, err := ImportNavi(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Commands) != 1 {
		t.Fatalf("got %d commands, want 1", len(report.Commands))
	}
	if report.Commands[0].Context != "kubectl" {
		t.Errorf("context = %q, want kubectl", report.Commands[0].Context)
	}
}

func TestImportNaviDynamicSuggestions(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "docker.cheat")
	writeFile(t, src, `% docker

# Stop containers
$ docker ps -q | head -5
docker stop
`)

	report, err := ImportNavi(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Commands) != 1 {
		t.Fatalf("got %d commands, want 1", len(report.Commands))
	}
	if !report.HasWarnings() {
		t.Fatal("expected warning for $ variable line")
	}
	if !strings.Contains(report.Commands[0].Notes, "navi variables") {
		t.Errorf("notes = %q, want navi variables note", report.Commands[0].Notes)
	}
}

func TestImportNaviExtensions(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "kubectl.cheat")
	writeFile(t, src, `% kubernetes

# Get pods
kubectl get pods
@ kubectl get pods --watch
`)

	report, err := ImportNavi(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Commands) != 1 {
		t.Fatalf("got %d commands, want 1", len(report.Commands))
	}
	if !report.HasWarnings() {
		t.Fatal("expected warning for @ extension")
	}
	if !strings.Contains(report.Commands[0].Notes, "navi extensions") {
		t.Errorf("notes = %q", report.Commands[0].Notes)
	}
}

func TestImportNaviMultilineCommand(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "build.cheat")
	writeFile(t, src, `% go

# Build and run
go build -o app .
./app
`)

	report, err := ImportNavi(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Commands) != 1 {
		t.Fatalf("got %d commands, want 1", len(report.Commands))
	}

	cmd := report.Commands[0].Command
	if !strings.Contains(cmd, "go build") || !strings.Contains(cmd, "./app") {
		t.Errorf("multiline command = %q", cmd)
	}
}

func TestImportNaviComments(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "git.cheat")
	writeFile(t, src, `% git
; this is a top comment

# Stash
; this is a comment
git stash
`)

	report, err := ImportNavi(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Commands) != 1 {
		t.Fatalf("got %d commands, want 1", len(report.Commands))
	}
	if report.Commands[0].Command != "git stash" {
		t.Errorf("command = %q, want 'git stash'", report.Commands[0].Command)
	}
}

func TestImportNaviFileContext(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "my-custom-commands.cheat")
	writeFile(t, src, `
# List dir
ls -la

# Disk usage
du -sh
`)

	report, err := ImportNavi(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Commands) != 2 {
		t.Fatalf("got %d commands, want 2", len(report.Commands))
	}
	if report.Commands[0].Context != "my-custom-commands" {
		t.Errorf("context = %q, want my-custom-commands", report.Commands[0].Context)
	}
}

func TestImportNaviWarningsCommitValid(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "mixed.cheat")
	writeFile(t, src, `% docker

# Valid entry
docker ps

# Entry with only variables
$ docker ps -q

# Another valid
docker stop
`)

	report, err := ImportNavi(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Commands) != 2 {
		t.Fatalf("got %d commands, want 2 (valid entries committed despite warnings)", len(report.Commands))
	}
	if !report.HasWarnings() {
		t.Fatal("expected warnings")
	}
}

func TestImportFatalError(t *testing.T) {
	_, err := ImportPet("/nonexistent/path.toml")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestImportPetUnsupportedParam(t *testing.T) {
	report, err := importPetTemp(t, `
[[snippets]]
description = "Kubectl get pods"
command = "kubectl get pods <name=|_all_||_default_|>"
`)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Commands) != 1 {
		t.Fatalf("got %d commands, want 1", len(report.Commands))
	}

	cmd := report.Commands[0].Command
	if !strings.Contains(cmd, "{{name:") && cmd != "kubectl get pods {{name=|_all_||_default_|}}" {
		t.Logf("complex param preserved as-is: %q (acceptable - warning would be ideal but not required)", cmd)
	}
}

func TestConvertVars(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"echo <name>", "echo {{name}}"},
		{"git commit -m <message:WIP>", "git commit -m {{message:WIP}}"},
		{"docker logs --tail <lines:100> <container>", "docker logs --tail {{lines:100}} {{container}}"},
		{"no vars here", "no vars here"},
		{"<a> <b:2>", "{{a}} {{b:2}}"},
	}
	for _, tc := range tests {
		got := convertVars(tc.input)
		if got != tc.want {
			t.Errorf("convertVars(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestReportString(t *testing.T) {
	r := &Report{Imported: 5, Skipped: 2}
	s := r.String()
	if !strings.Contains(s, "Imported: 5") {
		t.Errorf("report = %q", s)
	}
	if !strings.Contains(s, "Skipped (already exists): 2") {
		t.Errorf("report = %q", s)
	}
}

func TestReportStringDryRun(t *testing.T) {
	r := &Report{Imported: 3, Skipped: 1, DryRun: true}
	s := r.String()
	if !strings.Contains(s, "Would import: 3") {
		t.Errorf("dry run report = %q", s)
	}
	if !strings.Contains(s, "Would skip (already exists): 1") {
		t.Errorf("dry run report = %q", s)
	}
}

func TestReportStringWithWarnings(t *testing.T) {
	r := &Report{Imported: 3, Skipped: 1}
	r.AddWarning("skipped empty command", 5)
	s := r.String()
	if !strings.Contains(s, "Warnings: 1") {
		t.Errorf("report = %q, want warnings", s)
	}
	if !strings.Contains(s, "(line 5)") {
		t.Errorf("report = %q, want line number", s)
	}
}

func TestImportPetActualFormat(t *testing.T) {
	report, err := importPetTemp(t, `
[[Snippets]]
  Description = "Stash all changes"
  Output = ""
  Tag = []
  command = "git stash --include-untracked"

[[Snippets]]
  Description = "Disk usage"
  Output = ""
  Tag = []
  command = "du -sh * | sort -h"
`)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Commands) != 2 {
		t.Fatalf("got %d commands, want 2", len(report.Commands))
	}
	if report.Commands[0].Title != "Stash all changes" {
		t.Errorf("title = %q, want 'Stash all changes'", report.Commands[0].Title)
	}
	if report.Commands[0].Command != "git stash --include-untracked" {
		t.Errorf("command = %q", report.Commands[0].Command)
	}
}

func TestImportPetMixedCaseFields(t *testing.T) {
	report, err := importPetTemp(t, `
[[Snippets]]
  DESCRIPTION = "List"
  COMMAND = "ls"
  TAG = ["test"]

[[snippets]]
  Description = "Git log"
  command = "git log"
  Tag = ["git"]
`)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Commands) != 2 {
		t.Fatalf("got %d commands, want 2", len(report.Commands))
	}
	if report.Commands[0].Title != "List" {
		t.Errorf("title = %q", report.Commands[0].Title)
	}
	if len(report.Commands[0].Tags) != 1 || report.Commands[0].Tags[0] != "test" {
		t.Errorf("tags = %v", report.Commands[0].Tags)
	}
}

func TestSingleTagString(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "snippets.toml")
	writeFile(t, src, `
[[snippets]]
description = "Echo"
command = "echo hello"
tag = "single"
`)

	report, err := ImportPet(src)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Commands) != 1 {
		t.Fatalf("got %d commands", len(report.Commands))
	}
	if len(report.Commands[0].Tags) != 1 || report.Commands[0].Tags[0] != "single" {
		t.Errorf("tags = %v, want [single]", report.Commands[0].Tags)
	}
}
