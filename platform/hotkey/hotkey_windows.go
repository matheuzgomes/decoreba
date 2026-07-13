package hotkey

/*
#cgo LDFLAGS: -luser32
#include <windows.h>
#include <stdlib.h>
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

var keyMap = map[string]uintptr{
	"space":        0x20, // VK_SPACE
	"d":            0x44,
	"tab":          0x09,
	"escape":       0x1B,
	"backspace":    0x08,
	"return":       0x0D,
	"enter":        0x0D,
	"slash":        0xBF,
	"backslash":    0xDC,
	"period":       0xBE,
	"comma":        0xBC,
	"semicolon":    0xBA,
	"quoteleft":    0xC0,
	"grave":        0xC0,
	"minus":        0xBD,
	"equal":        0xBB,
	"bracketleft":  0xDB,
	"bracketright": 0xDD,
}

func newPlatform(showCh chan<- bool, key string) (*Manager, error) {
	vk, ok := keyMap[key]
	if !ok {
		return nil, fmt.Errorf("unknown key: %q (valid: %s)", key, KnownKeys())
	}

	mods := uintptr(C.MOD_ALT | C.MOD_SHIFT | C.MOD_NOREPEAT)
	kid := int32(1)

	ret := C.RegisterHotKey(nil, C.int(kid), C.uint(mods), C.uint(vk))
	if ret == 0 {
		return nil, fmt.Errorf("RegisterHotKey failed (key may be in use)")
	}

	log.Printf("hotkey: registered Alt+Shift+%s (id=%d)", key, kid)

	cancel := make(chan struct{})

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		for {
			select {
			case <-cancel:
				return
			default:
			}

			var msg C.MSG
			ret := C.PeekMessage(&msg, nil, C.WM_HOTKEY, C.WM_HOTKEY, C.PM_REMOVE)
			if ret != 0 && msg.message == C.WM_HOTKEY {
				log.Printf("hotkey: pressed!")
				select {
				case showCh <- true:
				default:
				}
			}

			C.Sleep(50)
		}
	}()

	return &Manager{closeFn: func() {
		close(cancel)
		C.UnregisterHotKey(nil, C.int(kid))
	}}, nil
}
