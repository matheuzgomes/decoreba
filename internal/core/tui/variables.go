package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/matheuzgomes/decoreba/internal/core/term"
)

// hasVariables reports whether cmd contains at least one {{...}} placeholder.
func hasVariables(cmd string) bool {
	return strings.Contains(cmd, "{{") && strings.Contains(cmd, "}}")
}

// resolveCommand prompts the user for each {{name:default}} variable in cmd
// and returns the final command string with all placeholders replaced.
// Returns the original command unchanged when there are no variables.
func resolveCommand(cmd string) (resolved string, cancelled bool, err error) {
	if !hasVariables(cmd) {
		return cmd, false, nil
	}

	result := cmd
	for {
		start := strings.Index(result, "{{")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "}}")
		if end == -1 {
			break
		}
		end += start

		inner := result[start+2 : end]
		name, def := inner, ""
		if colon := strings.Index(inner, ":"); colon >= 0 {
			name = inner[:colon]
			def = inner[colon+1:]
		}
		name = strings.TrimSpace(name)

		value, cancelled, err := promptVar(name, def)
		if err != nil {
			return "", false, err
		}
		if cancelled {
			return "", true, nil
		}

		if value == "" && def == "" {
			// No default and user left it empty: keep the placeholder literal.
			value = "{{" + inner + "}}"
		}
		result = result[:start] + value + result[end+2:]
	}
	return result, false, nil
}

// promptVar shows a single-line prompt for a variable and reads the value
// in raw mode. Returns the value, whether the user cancelled (Esc), and any
// error.
func promptVar(name, def string) (string, bool, error) {
	prompt := name
	if def != "" {
		prompt += " [" + def + "]"
	}
	fmt.Printf("\r\033[K%s: ", prompt)

	var buf []rune
	for {
		var b [1]byte
		n, err := os.Stdin.Read(b[:])
		if err != nil {
			return "", false, err
		}
		if n == 0 {
			continue
		}

		switch b[0] {
		case '\r', '\n':
			value := string(buf)
			if value == "" && def != "" {
				value = def
			}
			fmt.Print("\r\033[K")
			return value, false, nil
		case 0x1b:
			// Check for arrow sequences — consume and ignore.
			if term.InputAvailable(25) {
				extra := make([]byte, 8)
				os.Stdin.Read(extra)
			}
			fmt.Print("\r\033[K")
			return "", true, nil
		case 0x03:
			fmt.Print("\r\033[K")
			return "", true, nil
		case 0x7f, 0x08:
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				fmt.Print("\b \b")
			}
		default:
			if b[0] >= 0x20 {
				buf = append(buf, rune(b[0]))
				fmt.Print(string(b[0]))
			}
		}
	}
}
