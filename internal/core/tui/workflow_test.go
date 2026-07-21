package tui

import (
	"strings"
	"testing"

	"github.com/matheuzgomes/decoreba/internal/core"
)

func TestWorkflowRunner_FrameLines(t *testing.T) {
	w := &workflowRunner{
		cmd: &core.Command{
			Title: "Test Workflow",
			Steps: []core.WorkflowStep{
				{Title: "Step 1", Command: "echo 1"},
				{Title: "Step 2", Command: "echo 2"},
				{Title: "Step 3", Command: "echo 3"},
			},
		},
	}
	if got := w.frameLines(); got != 2+3*2+2 {
		t.Fatalf("frameLines = %d, want %d", got, 2+3*2+2)
	}
}

func TestWorkflowRunner_RenderContent(t *testing.T) {
	w := &workflowRunner{
		cmd: &core.Command{
			Title: "Deploy",
			Steps: []core.WorkflowStep{
				{Title: "Test", Command: "go test ./..."},
				{Title: "Build", Command: "go build"},
			},
		},
		status: []rune{'←', ' '},
		step:   0,
	}
	w.status[0] = '→'
	w.init(nil)
	w.width = 80
	w.height = 24

	body, cursorLine, cursorCol := w.renderContent(80, 24)
	if cursorLine != 1 || cursorCol != 0 {
		t.Fatalf("cursor = %d,%d", cursorLine, cursorCol)
	}
	lines := strings.Split(string(body), "\n")
	if len(lines) != 2+2*2+2 {
		t.Fatalf("got %d lines, want %d", len(lines), 2+2*2+2)
	}
	if !strings.Contains(string(body), "Deploy") {
		t.Fatalf("missing title: %s", string(body))
	}
	if !strings.Contains(string(body), "Test") || !strings.Contains(string(body), "Build") {
		t.Fatalf("missing steps: %s", string(body))
	}
	if !strings.Contains(string(body), "go test ./...") {
		t.Fatalf("missing command text: %s", string(body))
	}
}

func TestWorkflowRunner_ConfirmRender(t *testing.T) {
	w := &workflowRunner{
		cmd: &core.Command{
			Title: "Deploy",
			Steps: []core.WorkflowStep{
				{Title: "Test", Command: "go test ./..."},
			},
		},
		status:      []rune{'→'},
		step:        0,
		confirmExec: true,
	}
	w.init(nil)
	w.width = 80
	w.height = 24

	body, _, _ := w.renderContent(80, 24)
	if !strings.Contains(string(body), "Run all remaining steps") {
		t.Fatalf("confirm hint missing: %s", string(body))
	}
}

func TestWorkflowRunner_SuccessState(t *testing.T) {
	w := &workflowRunner{
		cmd: &core.Command{
			Title: "Deploy",
			Steps: []core.WorkflowStep{
				{Title: "Test", Command: "go test"},
			},
		},
		status: []rune{'✓'},
		step:   1,
	}
	w.init(nil)
	w.width = 80
	w.height = 24

	body, _, _ := w.renderContent(80, 24)
	if !strings.Contains(string(body), "✓") {
		t.Fatalf("success indicator missing: %s", string(body))
	}
}

func TestWorkflowRunner_FailState(t *testing.T) {
	w := &workflowRunner{
		cmd: &core.Command{
			Title: "Deploy",
			Steps: []core.WorkflowStep{
				{Title: "Test", Command: "go test"},
			},
		},
		status: []rune{'✗'},
		step:   1,
	}
	w.init(nil)
	w.width = 80
	w.height = 24

	body, _, _ := w.renderContent(80, 24)
	if !strings.Contains(string(body), "✗") {
		t.Fatalf("fail indicator missing: %s", string(body))
	}
}

func TestWorkflowRunner_NoSteps(t *testing.T) {
	// RunWorkflow returns nil early when called on a non-workflow command.
	cmd := &core.Command{Title: "Simple", Command: "echo hi"}
	if err := RunWorkflow(cmd); err != nil {
		t.Fatalf("RunWorkflow on non-workflow: %v", err)
	}
}
