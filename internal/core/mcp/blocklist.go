package mcp

import (
	"strings"
)

type BlocklistEntry struct {
	Substring string
	Label     string
	Severity  int
}

var blocklist = []BlocklistEntry{
	{`rm -rf /`, "rm -rf / (root filesystem)", 2},
	{`rm -rf ~`, "rm -rf ~ (home directory)", 2},
	{`rm -rf .`, "rm -rf . (current directory)", 1},
	{`rm -rf *`, "rm -rf * (all files)", 1},
	{`dd if=`, "dd if= (raw disk write)", 2},
	{`mkfs`, "mkfs (format filesystem)", 2},
	{`fdisk`, "fdisk (partition editor)", 2},
	{`format`, "format (disk format)", 2},
	{`> /dev/sd`, "write to block device", 2},
	{`:(){ :|:& };:`, "fork bomb", 2},
	{`chmod -R 0`, "chmod -R 0 (lock all files)", 1},
	{`chown -R`, "chown -R (recursive ownership)", 1},
	{`mv /`, "mv / (move root)", 2},
}

func CheckBlocklist(cmd string) (BlocklistEntry, bool) {
	lower := strings.ToLower(cmd)
	for _, e := range blocklist {
		if strings.Contains(lower, e.Substring) {
			return e, true
		}
	}
	return BlocklistEntry{}, false
}
