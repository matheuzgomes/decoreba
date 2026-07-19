package tui

import "unicode/utf8"

type keyKind int

const (
	keyRune keyKind = iota
	keyEnter
	keyEsc
	keyUp
	keyDown
	keyLeft
	keyRight
	keyBackspace
	keyDelete
	keyTab
	keyShiftTab
	keySave
	keyCancel
	keyEdit
	keyExecute
	keyWorkflow
	keyStepAdd
	keyStepDelete
)

type keyEvent struct {
	kind keyKind
	r    rune
}

func parseKeys(buf []byte) []keyEvent {
	var events []keyEvent
	for i := 0; i < len(buf); {
		b := buf[i]
		switch {
		case b == 0x1b:
			// CSI u: Shift+Enter → \x1b[13;2u
			if i+5 < len(buf) && buf[i+1] == '[' && buf[i+2] == '1' && buf[i+3] == '3' && buf[i+4] == ';' && buf[i+5] == '2' && i+6 < len(buf) && buf[i+6] == 'u' {
				events = append(events, keyEvent{kind: keyExecute})
				i += 7
				continue
			}

			if i+2 < len(buf) && (buf[i+1] == '[' || buf[i+1] == 'O') {
				switch buf[i+2] {
				case 'A':
					events = append(events, keyEvent{kind: keyUp})
					i += 3
					continue
				case 'B':
					events = append(events, keyEvent{kind: keyDown})
					i += 3
					continue
				case 'C':
					events = append(events, keyEvent{kind: keyRight})
					i += 3
					continue
				case 'D':
					events = append(events, keyEvent{kind: keyLeft})
					i += 3
					continue
				case 'Z':
					events = append(events, keyEvent{kind: keyShiftTab})
					i += 3
					continue
				}
			}
			if i+3 < len(buf) && buf[i+1] == '[' && buf[i+2] == '3' && buf[i+3] == '~' {
				events = append(events, keyEvent{kind: keyDelete})
				i += 4
				continue
			}
			events = append(events, keyEvent{kind: keyEsc})
			i++
		case b == '\r':
			events = append(events, keyEvent{kind: keyEnter})
			i++
		case b == 0x0a:
			events = append(events, keyEvent{kind: keyDown})
			i++
		case b == 0x0b:
			events = append(events, keyEvent{kind: keyUp})
			i++
		case b == 0x03:
			events = append(events, keyEvent{kind: keyCancel})
			i++
		case b == 0x05:
			events = append(events, keyEvent{kind: keyEdit})
			i++
		case b == 0x18:
			events = append(events, keyEvent{kind: keyExecute})
			i++
		case b == 0x17:
			events = append(events, keyEvent{kind: keyWorkflow})
			i++
		case b == 0x0E:
			events = append(events, keyEvent{kind: keyStepAdd})
			i++
		case b == 0x04:
			events = append(events, keyEvent{kind: keyStepDelete})
			i++
		case b == 0x09:
			events = append(events, keyEvent{kind: keyTab})
			i++
		case b == 0x13:
			events = append(events, keyEvent{kind: keySave})
			i++
		case b == 0x7f || b == 0x08:
			events = append(events, keyEvent{kind: keyBackspace})
			i++
		case b < 0x20:
			i++
		default:
			r, size := utf8.DecodeRune(buf[i:])
			if r == utf8.RuneError && size <= 1 {
				i++
				continue
			}
			events = append(events, keyEvent{kind: keyRune, r: r})
			i += size
		}
	}
	return events
}
