package tui

import (
	"reflect"
	"testing"
)

func TestParseKeys(t *testing.T) {
	r := func(r rune) keyEvent { return keyEvent{kind: keyRune, r: r} }
	k := func(kind keyKind) keyEvent { return keyEvent{kind: kind} }

	tests := []struct {
		name string
		in   string
		want []keyEvent
	}{
		{"arrow up", "\x1b[A", []keyEvent{k(keyUp)}},
		{"arrow down", "\x1b[B", []keyEvent{k(keyDown)}},
		{"arrow right", "\x1b[C", []keyEvent{k(keyRight)}},
		{"arrow left", "\x1b[D", []keyEvent{k(keyLeft)}},
		{"application cursor up", "\x1bOA", []keyEvent{k(keyUp)}},
		{"application cursor down", "\x1bOB", []keyEvent{k(keyDown)}},
		{"application cursor right", "\x1bOC", []keyEvent{k(keyRight)}},
		{"application cursor left", "\x1bOD", []keyEvent{k(keyLeft)}},
		{"shift+tab", "\x1b[Z", []keyEvent{k(keyShiftTab)}},
		{"delete", "\x1b[3~", []keyEvent{k(keyDelete)}},
		{"tab", "\t", []keyEvent{k(keyTab)}},
		{"ctrl+e", "\x05", []keyEvent{k(keyEdit)}},
		{"ctrl+x", "\x18", []keyEvent{k(keyExecute)}},
		{"shift+enter csi u", "\x1b[13;2u", []keyEvent{k(keyExecute)}},
		{"ctrl+s", "\x13", []keyEvent{k(keySave)}},
		{"lone esc", "\x1b", []keyEvent{k(keyEsc)}},
		{"printable ascii", "a", []keyEvent{r('a')}},
		{"utf8 rune", "ç", []keyEvent{r('ç')}},
		{"enter", "\r", []keyEvent{k(keyEnter)}},
		{"ctrl+c", "\x03", []keyEvent{k(keyCancel)}},
		{"ctrl+k is up", "\x0b", []keyEvent{k(keyUp)}},
		{"ctrl+j is down", "\x0a", []keyEvent{k(keyDown)}},
		{"backspace del", "\x7f", []keyEvent{k(keyBackspace)}},
		{"backspace bs", "\x08", []keyEvent{k(keyBackspace)}},
		{"ignored control byte", "\x01", nil},
		{"unknown escape sequence", "\x1b[H", []keyEvent{k(keyEsc), r('['), r('H')}},
		{
			"mixed chunk",
			"ab\x1b[A\r",
			[]keyEvent{r('a'), r('b'), k(keyUp), k(keyEnter)},
		},
		{
			"paste of several runes",
			"git ç\r",
			[]keyEvent{r('g'), r('i'), r('t'), r(' '), r('ç'), k(keyEnter)},
		},
		{"truncated utf8", "\xc3", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseKeys([]byte(tt.in))
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseKeys(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
