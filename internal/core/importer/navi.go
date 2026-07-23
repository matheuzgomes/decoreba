package importer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/matheuzgomes/decoreba/internal/core"
)

func ImportNavi(path string) (*Report, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("reading navi path: %w", err)
	}

	report := &Report{}
	now := time.Now()

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("reading navi directory: %w", err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasSuffix(name, ".cheat") && !strings.HasSuffix(name, ".cheat.md") {
				continue
			}
			err := parseNaviFile(filepath.Join(path, name), report, now)
			if err != nil {
				report.AddWarning(fmt.Sprintf("skipping %s: %v", name, err), 0)
			}
		}
	} else {
		err := parseNaviFile(path, report, now)
		if err != nil {
			return nil, fmt.Errorf("parsing navi file %s: %w", path, err)
		}
	}

	return report, nil
}

type naviBlock struct {
	tags    []string
	entries []naviEntry
}

type naviEntry struct {
	title   string
	lines   []string
	varLines []string
	extLines []string
}

func parseNaviFile(path string, report *Report, now time.Time) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	filename := strings.TrimSuffix(filepath.Base(path), ".cheat")
	filename = strings.TrimSuffix(filename, ".cheat.md")
	filename = strings.TrimSuffix(filename, ".md")
	fileContext := strings.TrimSpace(filename)

	blocks := parseNaviBlocks(string(data))

	for _, block := range blocks {
		tags := block.tags
		if len(tags) == 0 && fileContext != "" {
			tags = []string{fileContext}
		}

		for _, entry := range block.entries {
			title := strings.TrimSpace(entry.title)
			if title == "" {
				title = fileContext
			}

			var cmdLines []string
			var notesParts []string

			for _, l := range entry.lines {
				trimmed := strings.TrimSpace(l)
				if trimmed == "" {
					continue
				}
				if strings.HasPrefix(trimmed, ";") {
					continue
				}
				if strings.HasPrefix(trimmed, "$") {
					entry.varLines = append(entry.varLines, trimmed)
					continue
				}
				if strings.HasPrefix(trimmed, "@") {
					entry.extLines = append(entry.extLines, trimmed)
					continue
				}
				cmdLines = append(cmdLines, trimmed)
			}

			if len(cmdLines) == 0 {
				report.AddWarning(fmt.Sprintf("skipped entry %q with empty command", title), 0)
				continue
			}

			if len(entry.varLines) > 0 {
				notesParts = append(notesParts, "navi variables: "+strings.Join(entry.varLines, "; "))
				report.AddWarning(fmt.Sprintf("entry %q has $ dynamic suggestions that were not imported", title), 0)
			}
			if len(entry.extLines) > 0 {
				notesParts = append(notesParts, "navi extensions: "+strings.Join(entry.extLines, "; "))
				report.AddWarning(fmt.Sprintf("entry %q has @ extensions that were not imported", title), 0)
			}

			notes := strings.Join(notesParts, "\n")

			cmd := core.Command{
				ID:        core.GenID(),
				Context:   fileContext,
				Title:     title,
				Command:   convertVars(strings.Join(cmdLines, "\n")),
				Tags:      tags,
				Notes:     notes,
				CreatedAt: now,
				UpdatedAt: now,
			}

			report.Commands = append(report.Commands, cmd)
		}
	}

	return nil
}

func parseNaviBlocks(input string) []naviBlock {
	var blocks []naviBlock
	lines := strings.Split(input, "\n")

	var current *naviBlock
	var currentEntry *naviEntry

	flushEntry := func() {
		if currentEntry != nil && (len(currentEntry.lines) > 0 || currentEntry.title != "") {
			current.entries = append(current.entries, *currentEntry)
		}
		currentEntry = nil
	}

	for _, raw := range lines {
		line := strings.TrimRightFunc(raw, unicode.IsSpace)

		if strings.HasPrefix(line, "%") {
			flushEntry()
			if current != nil && len(current.entries) > 0 {
				blocks = append(blocks, *current)
			}

			tagStr := strings.TrimSpace(line[1:])
			var tags []string
			for _, t := range strings.Split(tagStr, ",") {
				tt := strings.TrimSpace(t)
				if tt != "" {
					tags = append(tags, tt)
				}
			}

			current = &naviBlock{tags: tags}
			currentEntry = &naviEntry{}
			continue
		}

		if current == nil {
			current = &naviBlock{}
			currentEntry = &naviEntry{}
		}

		if strings.HasPrefix(line, "#") {
			flushEntry()
			currentEntry = &naviEntry{title: strings.TrimSpace(line[1:])}
			continue
		}

		if currentEntry == nil {
			currentEntry = &naviEntry{}
		}

		currentEntry.lines = append(currentEntry.lines, line)
	}

	flushEntry()
	if current != nil && len(current.entries) > 0 {
		blocks = append(blocks, *current)
	}

	return blocks
}
