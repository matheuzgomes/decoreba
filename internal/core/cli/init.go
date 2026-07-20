package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const blockBegin = "# --- decoreba begin ---"
const blockEnd = "# --- decoreba end ---"

func cmdInit(args []string) {
	dryRun := false
	autoYes := false
	forceShell := ""

	for _, a := range args {
		switch a {
		case "--dry-run":
			dryRun = true
		case "-y", "--yes":
			autoYes = true
		case "bash", "zsh", "fish":
			forceShell = a
		}
	}

	shell := forceShell
	if shell == "" {
		shell = os.Getenv("SHELL")
	}
	switch {
	case strings.HasSuffix(shell, "zsh"):
		shell = "zsh"
	case strings.HasSuffix(shell, "fish"):
		shell = "fish"
	default:
		shell = "bash"
	}

	rcFile := rcFilePath(shell)
	if rcFile == "" {
		fmt.Fprintf(os.Stderr, "error: no config file found for %s\n", shell)
		os.Exit(1)
	}

	completionCmd := fmt.Sprintf("source <(decoreba completion %s)", shell)
	var shellCmd string
	switch shell {
	case "bash":
		shellCmd = "source <(decoreba shell bash)"
	case "zsh":
		shellCmd = "source <(decoreba shell zsh)"
	default:
		shellCmd = ""
	}

	var lines []string
	lines = append(lines, blockBegin)
	lines = append(lines, completionCmd)
	if shellCmd != "" {
		lines = append(lines, shellCmd)
	}
	lines = append(lines, blockEnd)
	block := strings.Join(lines, "\n") + "\n"

	if dryRun {
		fmt.Printf("# Shell: %s\n", shell)
		fmt.Printf("# Config: %s\n", rcFile)
		fmt.Println("# Block to install:")
		fmt.Println(block)
		return
	}

	existing, err := os.ReadFile(rcFile)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", rcFile, err)
		os.Exit(1)
	}

	content := string(existing)

	hasBlock := strings.Contains(content, blockBegin)

	if !autoYes {
		fmt.Printf("Shell:   %s\n", shell)
		fmt.Printf("Config:  %s\n", rcFile)
		if hasBlock {
			fmt.Println("Action:  update decoreba block")
		} else {
			fmt.Println("Action:  append decoreba block")
		}
		fmt.Print("Proceed? y/n: ")
		answer := strings.TrimSpace(promptLine(""))
		if answer != "y" && answer != "Y" && answer != "yes" {
			fmt.Println("Aborted.")
			return
		}
	}

	if hasBlock {
		start := strings.Index(content, blockBegin)
		end := strings.Index(content, blockEnd)
		if end >= 0 {
			end += len(blockEnd)
		}
		before := content[:start]
		after := ""
		if end < len(content) {
			after = content[end:]
		}
		content = before + block + strings.TrimPrefix(after, "\n")
	} else {
		rcDir := filepath.Dir(rcFile)
		if err := os.MkdirAll(rcDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "error creating %s: %v\n", rcDir, err)
			os.Exit(1)
		}
		if !strings.HasSuffix(content, "\n") && content != "" {
			content += "\n"
		}
		content += "\n" + block
	}

	backup := rcFile + "." + time.Now().Format("20060102150405") + ".bak"
	if err := os.WriteFile(backup, existing, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error backing up to %s: %v\n", backup, err)
		os.Exit(1)
	}

	if err := os.WriteFile(rcFile, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", rcFile, err)
		os.Exit(1)
	}

	fmt.Printf("✔ Backup: %s\n", backup)
	fmt.Printf("✔ Config: %s\n", rcFile)
	if shellCmd != "" {
		fmt.Println("✔ Widget: Ctrl+O")
	}
	fmt.Println("✔ Completions")
	if shell != "fish" {
		fmt.Printf("→ %s\n", sourceCmd(shell, rcFile))
	}
}

func rcFilePath(shell string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	switch shell {
	case "bash":
		if _, err := os.Stat(filepath.Join(home, ".bashrc")); err == nil {
			return filepath.Join(home, ".bashrc")
		}
		return filepath.Join(home, ".bash_profile")
	case "zsh":
		return filepath.Join(home, ".zshrc")
	case "fish":
		return filepath.Join(home, ".config", "fish", "config.fish")
	}
	return ""
}

func sourceCmd(shell, rcFile string) string {
	home, _ := os.UserHomeDir()
	rel := rcFile
	if home != "" {
		if r, err := filepath.Rel(home, rcFile); err == nil && len(r) < len(rel) {
			rel = "~/" + r
		}
	}
	switch shell {
	case "fish":
		return fmt.Sprintf("source %s", rel)
	default:
		return fmt.Sprintf("source %s", rel)
	}
}
