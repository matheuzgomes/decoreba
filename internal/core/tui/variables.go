package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/matheuzgomes/decoreba/internal/core/term"
)

// HasVariables reports whether cmd contains at least one {{...}} placeholder.
func HasVariables(cmd string) bool {
	return strings.Contains(cmd, "{{") && strings.Contains(cmd, "}}")
}

// ResolveCommandInteractive prompts the user for each {{name:default}} variable
// and returns the resolved command. Manages its own terminal raw mode.
func ResolveCommandInteractive(cmd string) (resolved string, cancelled bool, err error) {
	if !HasVariables(cmd) {
		return cmd, false, nil
	}

	vars := parseVars(cmd)
	if len(vars) == 0 {
		return cmd, false, nil
	}

	restore, err := term.MakeRaw()
	if err != nil {
		return "", false, err
	}
	defer restore()

	out := os.Stdout
	if UseTTY {
		if f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
			out = f
			defer f.Close()
		}
	}

	// Show a reference line so the user knows what they're filling in.
	fmt.Fprintf(out, "\r\x1b[K%s %s\n", ansiDim+"cmd:"+ansiReset, cmd)

	result := cmd
	for _, v := range vars {
		value, cancelled, err := readVar(out, v.name, v.def)
		if err != nil {
			return "", false, err
		}
		if cancelled {
			fmt.Fprint(out, "\r\x1b[K")
			return "", true, nil
		}
		result = strings.Replace(result, "{{"+v.raw+"}}", value, 1)
	}

	// Clear the reference line.
	fmt.Fprint(out, "\r\x1b[1A\x1b[K")
	return result, false, nil
}

// readVar reads a single variable value from the terminal. Raw mode must be
// active.
func readVar(out *os.File, name, def string) (string, bool, error) {
	prompt := name
	if def != "" {
		prompt += " [" + def + "]"
	}
	fmt.Fprintf(out, "\r\x1b[K%s: ", prompt)

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
			if value == "" && def == "" {
				value = "{{" + name + "}}"
			}
			fmt.Fprint(out, "\r\x1b[K")
			return value, false, nil
		case 0x1b:
			if term.InputAvailable(25) {
				extra := make([]byte, 8)
				os.Stdin.Read(extra)
			}
			fmt.Fprint(out, "\r\x1b[K")
			return "", true, nil
		case 0x03:
			fmt.Fprint(out, "\r\x1b[K")
			return "", true, nil
		case 0x7f, 0x08:
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				fmt.Fprint(out, "\b \b")
			}
		default:
			if b[0] >= 0x20 {
				buf = append(buf, rune(b[0]))
				fmt.Fprint(out, string(b[0]))
			}
		}
	}
}

func parseVars(cmd string) []varInfo {
	var vars []varInfo
	s := cmd
	for {
		start := strings.Index(s, "{{")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], "}}")
		if end == -1 {
			break
		}
		end += start
		inner := s[start+2 : end]
		name, def := inner, ""
		if colon := strings.Index(inner, ":"); colon >= 0 {
			name = strings.TrimSpace(inner[:colon])
			def = inner[colon+1:]
		} else {
			name = strings.TrimSpace(name)
		}
		vars = append(vars, varInfo{
			name: name,
			def:  def,
			raw:  inner,
		})
		s = s[end+2:]
	}
	return vars
}

type varInfo struct {
	name string
	def  string
	raw  string
}
