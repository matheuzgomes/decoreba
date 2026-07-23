package importer

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/matheuzgomes/decoreba/internal/core"
)

type petSnippet struct {
	Description string   `toml:"description"`
	Command     string   `toml:"command"`
	Tag         []string `toml:"tag"`
	Output      string   `toml:"output"`
	line        int
}

func ImportPet(path string) (*Report, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading pet file: %w", err)
	}

	snippets, err := parsePetTOML(string(data))
	if err != nil {
		return nil, fmt.Errorf("parsing pet file: %w", err)
	}

	report := &Report{}
	now := time.Now()

	for _, s := range snippets {
		if strings.TrimSpace(s.Command) == "" {
			report.AddWarning("skipped entry with empty command", s.line)
			continue
		}

		title := strings.TrimSpace(s.Description)
		if title == "" {
			firstLine := strings.SplitN(strings.TrimSpace(s.Command), "\n", 2)[0]
			title = strings.TrimSpace(firstLine)
		}

		notes := ""
		if s.Output != "" {
			notes = fmt.Sprintf("pet output: %s", s.Output)
		}

		cmd := core.Command{
			ID:        core.GenID(),
			Context:   "pet-imported",
			Title:     title,
			Command:   convertVars(strings.TrimSpace(s.Command)),
			Tags:      s.Tag,
			Notes:     notes,
			CreatedAt: now,
			UpdatedAt: now,
		}

		report.Commands = append(report.Commands, cmd)
	}

	return report, nil
}

func parsePetTOML(input string) ([]petSnippet, error) {
	var snippets []petSnippet
	lines := strings.Split(input, "\n")

	var current *petSnippet
	inSnippet := false
	inMultiline := false
	multilineKey := ""
	multilineBuf := strings.Builder{}

	for i, line := range lines {
		lineNum := i + 1

		if inMultiline {
			closer := strings.Index(line, `"""`)
			if closer != -1 {
				rest := line[:closer]
				if rest != "" {
					multilineBuf.WriteString("\n")
					multilineBuf.WriteString(rest)
				}
				val := strings.TrimSpace(multilineBuf.String())
				setPetField(current, multilineKey, val)
				inMultiline = false
				multilineKey = ""
			} else {
				multilineBuf.WriteString("\n")
				multilineBuf.WriteString(line)
			}
			continue
		}

		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.EqualFold(trimmed, "[[snippets]]") {
			if current != nil {
				snippets = append(snippets, *current)
			}
			current = &petSnippet{line: lineNum}
			inSnippet = true
			continue
		}

		if !inSnippet {
			continue
		}

		eqIdx := strings.Index(trimmed, "=")
		if eqIdx == -1 {
			continue
		}

		key := strings.TrimSpace(trimmed[:eqIdx])
		value := strings.TrimSpace(trimmed[eqIdx+1:])

		if strings.HasPrefix(value, `"""`) {
			rest := strings.TrimPrefix(value, `"""`)
			end := strings.Index(rest, `"""`)
			if end != -1 {
				val := strings.TrimSpace(rest[:end])
				setPetField(current, key, val)
			} else {
				inMultiline = true
				multilineKey = key
				multilineBuf.Reset()
				if rest != "" {
					multilineBuf.WriteString(rest)
				}
			}
			continue
		}

		switch strings.ToLower(key) {
		case "description":
			current.Description = unquoteTOML(value)
		case "command":
			current.Command = unquoteTOML(value)
		case "output":
			current.Output = unquoteTOML(value)
		case "tag":
			tags := parseTOMLArray(value)
			current.Tag = append(current.Tag, tags...)
		}
	}

	if current != nil {
		snippets = append(snippets, *current)
	}

	return snippets, nil
}

func setPetField(s *petSnippet, key, value string) {
	switch strings.ToLower(key) {
	case "description":
		s.Description = value
	case "command":
		s.Command = value
	case "output":
		s.Output = value
	case "tag":
		s.Tag = append(s.Tag, value)
	}
}

func unquoteTOML(s string) string {
	if len(s) >= 6 && s[:3] == `"""` && s[len(s)-3:] == `"""` {
		return s[3 : len(s)-3]
	}
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func parseTOMLArray(s string) []string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "[") || !strings.HasSuffix(s, "]") {
		if s != "" {
			return []string{unquoteTOML(s)}
		}
		return nil
	}

	inner := s[1 : len(s)-1]
	if strings.TrimSpace(inner) == "" {
		return nil
	}

	var items []string
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(inner); i++ {
		ch := inner[i]
		if inQuote {
			current.WriteByte(ch)
			if ch == quoteChar {
				inQuote = false
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			if current.Len() > 0 {
				items = append(items, strings.TrimSpace(current.String()))
				current.Reset()
			}
			inQuote = true
			quoteChar = ch
			current.WriteByte(ch)
			continue
		}
		if ch == ',' {
			items = append(items, strings.TrimSpace(current.String()))
			current.Reset()
			continue
		}
		if !unicode.IsSpace(rune(ch)) {
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		items = append(items, strings.TrimSpace(current.String()))
	}

	result := make([]string, 0, len(items))
	for _, item := range items {
		unquoted := unquoteTOML(item)
		if unquoted != "" {
			result = append(result, unquoted)
		}
	}

	return result
}
