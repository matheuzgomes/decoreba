package mcp

import (
	"testing"
)

func TestCheckBlocklist(t *testing.T) {
	tests := []struct {
		cmd     string
		blocked bool
		label   string
	}{
		{"docker container prune", false, ""},
		{"rm -rf /", true, "rm -rf / (root filesystem)"},
		{"rm -rf ~/tmp", true, "rm -rf ~ (home directory)"},
		{"rm -rf .", true, "rm -rf . (current directory)"},
		{"rm -rf *", true, "rm -rf * (all files)"},
		{"dd if=/dev/zero of=/dev/sda", true, "dd if= (raw disk write)"},
		{"mkfs.ext4 /dev/sdb1", true, "mkfs (format filesystem)"},
		{"fdisk /dev/sda", true, "fdisk (partition editor)"},
		{"format C:", true, "format (disk format)"},
		{"> /dev/sda", true, "write to block device"},
		{":(){ :|:& };:", true, "fork bomb"},
		{"chmod -R 0 /home", true, "chmod -R 0 (lock all files)"},
		{"chown -R user:group /var", true, "chown -R (recursive ownership)"},
		{"mv / /tmp/root", true, "mv / (move root)"},
	}

	for _, tt := range tests {
		entry, blocked := CheckBlocklist(tt.cmd)
		if blocked != tt.blocked {
			t.Errorf("CheckBlocklist(%q): blocked=%v, want %v", tt.cmd, blocked, tt.blocked)
		}
		if blocked && entry.Label != tt.label {
			t.Errorf("CheckBlocklist(%q): label=%q, want %q", tt.cmd, entry.Label, tt.label)
		}
	}
}

func TestCheckBlocklistCaseInsensitive(t *testing.T) {
	entry, blocked := CheckBlocklist("RM -RF /")
	if !blocked {
		t.Fatal("should be case-insensitive")
	}
	if entry.Severity != 2 {
		t.Fatalf("severity = %d, want 2", entry.Severity)
	}
}

func TestCheckBlocklistPartial(t *testing.T) {
	_, blocked := CheckBlocklist("rm -rf /var/log")
	if !blocked {
		t.Fatal("'rm -rf /var/log' should be blocked (contains 'rm -rf /')")
	}
}

func TestCheckBlocklistSeverity(t *testing.T) {
	t.Run("severity 2 for destructive", func(t *testing.T) {
		entry, ok := CheckBlocklist("rm -rf /")
		if !ok || entry.Severity != 2 {
			t.Fatalf("rm -rf /: severity=%d", entry.Severity)
		}
	})

	t.Run("severity 1 for moderate", func(t *testing.T) {
		entry, ok := CheckBlocklist("rm -rf .")
		if !ok || entry.Severity != 1 {
			t.Fatalf("rm -rf .: severity=%d", entry.Severity)
		}
	})
}
