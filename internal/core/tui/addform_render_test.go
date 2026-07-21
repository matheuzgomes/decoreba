package tui

import (
	"strings"
	"testing"

	"github.com/matheuzgomes/decoreba/internal/core"
)

func TestAddFormRenderEditMode(t *testing.T) {
	f := &addForm{
		store:    &core.Store{},
		errField: -1,
		editing:  true,
		existing: &core.Command{
			ID:      "abc123",
			Context: "git",
			Title:   "Undo",
			Command: "git reset",
		},
	}
	f.width = 80
	f.height = 24
	f.fields[fieldContext] = []rune("git")
	f.fields[fieldTitle] = []rune("Undo")
	f.fields[fieldCommand] = []rune("git reset")

	frame := string(f.renderFrame())
	if !strings.Contains(frame, editCmdHeader) {
		t.Fatalf("edit header missing: %q", frame)
	}
}

func TestAddFormRenderWorkflowMode(t *testing.T) {
	f := newTestAddForm()
	f.isWorkflow = true
	f.fields[fieldContext] = []rune("git")
	f.fields[fieldTitle] = []rune("Deploy")
	f.workflowSteps = []core.WorkflowStep{
		{Title: "Test", Command: "go test"},
	}

	frame := string(f.renderFrame())
	if !strings.Contains(frame, "steps") {
		t.Fatalf("should show workflow hint: %q", frame)
	}
}

func TestAddFormRenderErrorFlash(t *testing.T) {
	f := newTestAddForm()
	f.errField = fieldContext
	f.errMsg = "context is required"
	f.errFlash = 3

	frame := string(f.renderFrame())
	if !strings.Contains(frame, "context is required") {
		t.Fatalf("error message missing: %q", frame)
	}
}

func TestAddFormEmptyContext(t *testing.T) {
	f := newTestAddForm()
	// Test with empty fields to ensure it renders
	frame := string(f.renderFrame())
	if !strings.Contains(frame, newCmdHeader) {
		t.Fatalf("new command header missing: %q", frame)
	}
}

func TestAddFormConfirmExit(t *testing.T) {
	f := newTestAddForm()
	f.confirmExit = true
	f.errMsg = "Unsaved changes"
	f.errFlash = 4

	frame := string(f.renderFrame())
	if !strings.Contains(frame, "Unsaved changes") {
		t.Fatalf("confirm exit message missing: %q", frame)
	}
}
