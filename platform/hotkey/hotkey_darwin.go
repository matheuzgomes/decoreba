package hotkey

/*
#cgo LDFLAGS: -framework Carbon
#include <Carbon/Carbon.h>
#include <stdlib.h>

extern void goHotkeyCallback();

static EventHotKeyRef hotKeyRef = NULL;

static OSStatus hotkeyHandler(EventHandlerCallRef ref, EventRef event, void *data) {
	goHotkeyCallback();
	return noErr;
}

static int registerHotkey(unsigned int keycode, unsigned int modifiers) {
	EventTypeSpec spec;
	spec.eventClass = kEventClassKeyboard;
	spec.eventKind = kEventHotKeyPressed;

	EventTargetRef target = GetApplicationEventTarget();

	OSStatus err = InstallEventHandler(target, &hotkeyHandler, 1, &spec, NULL, NULL);
	if (err != noErr) return 0;

	EventHotKeyID hkID;
	hkID.signature = 'deco';
	hkID.id = 1;

	err = RegisterEventHotKey(keycode, modifiers, hkID, target, 0, &hotKeyRef);
	return err == noErr ? 1 : 0;
}

static void unregisterHotkey() {
	if (hotKeyRef != NULL) {
		UnregisterEventHotKey(hotKeyRef);
		hotKeyRef = NULL;
	}
}
*/
import "C"
import (
	"fmt"
	"log"
	"runtime"
)

var keyNames = map[string]struct{}{
	"space": {}, "d": {}, "tab": {}, "escape": {}, "backspace": {},
	"return": {}, "enter": {}, "slash": {}, "backslash": {},
	"period": {}, "comma": {}, "semicolon": {}, "quoteleft": {},
	"grave": {}, "minus": {}, "equal": {},
	"bracketleft": {}, "bracketright": {},
}

var keyMap = map[string]uint32{
	"space":        0x31, // kVK_Space
	"d":            0x02, // kVK_ANSI_D
	"tab":          0x30, // kVK_Tab
	"escape":       0x35, // kVK_Escape
	"backspace":    0x33, // kVK_Delete
	"return":       0x24, // kVK_Return
	"enter":        0x24,
	"slash":        0x2C, // kVK_ANSI_Slash
	"backslash":    0x2A, // kVK_ANSI_Backslash
	"period":       0x2F, // kVK_ANSI_Period
	"comma":        0x2B, // kVK_ANSI_Comma
	"semicolon":    0x29, // kVK_ANSI_Semicolon
	"quoteleft":    0x32, // kVK_ANSI_Grave
	"grave":        0x32,
	"minus":        0x1B, // kVK_ANSI_Minus
	"equal":        0x18, // kVK_ANSI_Equal
	"bracketleft":  0x21, // kVK_ANSI_LeftBracket
	"bracketright": 0x1E, // kVK_ANSI_RightBracket
}

var hotkeyShowCh chan<- bool

const (
	macCmdKey   = 1 << 8 // cmdKey bit in Carbon
	macShiftKey = 1 << 9 // shiftKey bit in Carbon
)

func newPlatform(showCh chan<- bool, key string) (*Manager, error) {
	kv, ok := keyMap[key]
	if !ok {
		return nil, fmt.Errorf("unknown key: %q (valid: %s)", key, KnownKeys())
	}

	hotkeyShowCh = showCh

	mods := C.uint(macCmdKey | macShiftKey)

	ret := C.registerHotkey(C.uint(kv), mods)
	if ret == 0 {
		return nil, fmt.Errorf("RegisterEventHotKey failed")
	}

	log.Printf("hotkey: registered Cmd+Shift+%s (macOS keycode=%d)", key, kv)

	cancel := make(chan struct{})

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		log.Printf("hotkey: event loop started, waiting for Cmd+Shift+%s", key)

		runLoop := C.CFRunLoopGetCurrent()
		for {
			select {
			case <-cancel:
				log.Printf("hotkey: event loop cancelled")
				C.CFRunLoopStop(runLoop)
				return
			default:
				C.CFRunLoopRunInMode(C.kCFRunLoopDefaultMode, 0.1, 1)
			}
		}
	}()

	return &Manager{closeFn: func() {
		close(cancel)
		C.unregisterHotkey()
	}}, nil
}

//export goHotkeyCallback
func goHotkeyCallback() {
	select {
	case hotkeyShowCh <- true:
	default:
	}
}
