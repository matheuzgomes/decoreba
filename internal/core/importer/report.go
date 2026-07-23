package importer

import (
	"fmt"
	"strings"

	"github.com/matheuzgomes/decoreba/internal/core"
)

type Warning struct {
	Message string
	Line    int
}

type Report struct {
	Imported int
	Skipped  int
	Warnings []Warning
	Commands []core.Command
	DryRun   bool
}

func (r *Report) AddWarning(msg string, line int) {
	r.Warnings = append(r.Warnings, Warning{Message: msg, Line: line})
}

func (r *Report) HasWarnings() bool {
	return len(r.Warnings) > 0
}

func (r *Report) String() string {
	var b strings.Builder
	importedLabel := "Imported"
	skippedLabel := "Skipped (already exists)"
	if r.DryRun {
		importedLabel = "Would import"
		skippedLabel = "Would skip (already exists)"
	}
	fmt.Fprintf(&b, "%s: %d\n", importedLabel, r.Imported)
	fmt.Fprintf(&b, "%s: %d\n", skippedLabel, r.Skipped)
	if r.HasWarnings() {
		fmt.Fprintf(&b, "Warnings: %d\n", len(r.Warnings))
		for _, w := range r.Warnings {
			line := ""
			if w.Line > 0 {
				line = fmt.Sprintf(" (line %d)", w.Line)
			}
			fmt.Fprintf(&b, "  - %s%s\n", w.Message, line)
		}
	}
	return b.String()
}

func convertVars(s string) string {
	result := s
	for {
		start := strings.Index(result, "<")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], ">")
		if end == -1 {
			break
		}
		end += start
		inner := result[start+1 : end]
		eq := strings.Index(inner, "=")
		if eq != -1 {
			name := inner[:eq]
			def := inner[eq+1:]
			result = result[:start] + "{{" + name + ":" + def + "}}" + result[end+1:]
		} else {
			result = result[:start] + "{{" + inner + "}}" + result[end+1:]
		}
	}
	return result
}
