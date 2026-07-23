package history

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var errUnsupportedShell = errors.New("unsupported shell; use bash, zsh, or fish")

type Entry struct {
	Command string
}

func Last() (string, error) {
	entries, err := readEntries()
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("no command found in history file")
	}
	return entries[len(entries)-1].Command, nil
}

func LastExcludingSelf() (string, error) {
	entries, err := readEntries()
	if err != nil {
		return "", err
	}

	for i := len(entries) - 1; i >= 0; i-- {
		if !isSelfCall(entries[i].Command) {
			return entries[i].Command, nil
		}
	}
	return "", fmt.Errorf("no command found before decoreba invocation")
}

func isSelfCall(cmd string) bool {
	parts := strings.Fields(cmd)
	if len(parts) < 3 {
		return false
	}
	executable := filepath.Base(parts[0])
	return executable == "decoreba" && parts[1] == "add" && parts[2] == "--last"
}

func readEntries() ([]Entry, error) {
	shell := filepath.Base(os.Getenv("SHELL"))
	if shell != "bash" && shell != "zsh" && shell != "fish" {
		return nil, errUnsupportedShell
	}

	path := os.Getenv("HISTFILE")
	if path == "" {
		if shell == "fish" {
			dir := os.Getenv("XDG_DATA_HOME")
			if dir == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					return nil, err
				}
				dir = filepath.Join(home, ".local", "share")
			}
			name := os.Getenv("fish_history")
			if name == "" || name == "default" {
				name = "fish"
			}
			path = filepath.Join(dir, "fish", name+"_history")
		}
	}
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		name := ".bash_history"
		if shell == "zsh" {
			name = ".zsh_history"
		}
		path = filepath.Join(home, name)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s history file %q: %w", shell, path, err)
	}

	entries := parseAll(string(data), shell)
	if len(entries) == 0 {
		return nil, fmt.Errorf("no command found in %s history file", shell)
	}
	return entries, nil
}

func parseAll(data, shell string) []Entry {
	if shell == "fish" {
		return parseFishAll(data)
	}

	lines := strings.Split(strings.ReplaceAll(data, "\r\n", "\n"), "\n")
	var entries []Entry

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if shell == "zsh" && strings.HasPrefix(line, ": ") {
			var cmdParts []string
			parts := strings.SplitN(line, ";", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
				cmdParts = append(cmdParts, parts[1])
			}
			i++
			for i < len(lines) {
				next := lines[i]
				if strings.HasPrefix(next, ": ") {
					i--
					break
				}
				if strings.TrimSpace(next) != "" {
					cmdParts = append(cmdParts, next)
				}
				i++
			}
			if len(cmdParts) > 0 {
				entries = append(entries, Entry{Command: strings.TrimSpace(strings.Join(cmdParts, "\n"))})
			}
		} else if shell == "bash" && strings.HasPrefix(line, "#") && isDigits(line[1:]) {
			i++
			var cmdParts []string
			for i < len(lines) {
				next := lines[i]
				if strings.HasPrefix(next, "#") && isDigits(next[1:]) {
					i--
					break
				}
				if strings.TrimSpace(next) != "" {
					cmdParts = append(cmdParts, next)
				}
				i++
			}
			if len(cmdParts) > 0 {
				entries = append(entries, Entry{Command: strings.TrimSpace(strings.Join(cmdParts, "\n"))})
			}
		}
	}

	if len(entries) > 0 {
		return entries
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			entries = append(entries, Entry{Command: trimmed})
		}
	}
	return entries
}

func parseFishAll(data string) []Entry {
	var entries []Entry
	for _, line := range strings.Split(strings.ReplaceAll(data, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- cmd: ") {
			cmd := decodeValue(strings.TrimPrefix(line, "- cmd: "))
			if cmd != "" {
				entries = append(entries, Entry{Command: strings.TrimSpace(cmd)})
			}
		}
	}
	return entries
}

func parse(data, shell string) string {
	entries := parseAll(data, shell)
	if len(entries) > 0 {
		return entries[len(entries)-1].Command
	}
	return ""
}

func decodeValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
		return strings.ReplaceAll(value[1:len(value)-1], "''", "'")
	}
	if len(value) >= 2 && value[0] == '"' {
		if decoded, err := strconv.Unquote(value); err == nil {
			return decoded
		}
	}
	return value
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
