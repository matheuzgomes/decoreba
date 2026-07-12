package hotkey

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/keysym.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unsafe"
)

type Manager struct {
	display *C.Display
	root    C.Window
	showCh  chan<- bool
	cancel  chan struct{}
}

var keysymMap = map[string]C.KeySym{
	"space":      C.XK_space,
	"d":          C.XK_D,
	"tab":        C.XK_Tab,
	"escape":     C.XK_Escape,
	"backspace":  C.XK_BackSpace,
	"return":     C.XK_Return,
	"enter":      C.XK_Return,
	"slash":      C.XK_slash,
	"backslash":  C.XK_backslash,
	"period":     C.XK_period,
	"comma":      C.XK_comma,
	"semicolon":  C.XK_semicolon,
	"quoteleft":  C.XK_quoteleft,
	"grave":      C.XK_grave,
	"minus":      C.XK_minus,
	"equal":      C.XK_equal,
	"bracketleft":  C.XK_bracketleft,
	"bracketright": C.XK_bracketright,
}

func New(showCh chan<- bool) (*Manager, error) {
	return NewKey(showCh, "space")
}

func NewKey(showCh chan<- bool, key string) (*Manager, error) {
	keysym, ok := keysymMap[strings.ToLower(key)]
	if !ok {
		return nil, fmt.Errorf("unknown key: %q (valid: %s)", key, knownKeys())
	}

	if os.Getenv("WAYLAND_DISPLAY") != "" {
		log.Printf("hotkey: Wayland detected — global hotkeys may not work via X11.")
	}

	cs := C.CString("")
	defer C.free(unsafe.Pointer(cs))
	dpy := C.XOpenDisplay(cs)
	if dpy == nil {
		return nil, fmt.Errorf("cannot open X display — is DISPLAY set?")
	}
	log.Printf("hotkey: X11 display opened (DISPLAY=%s)", os.Getenv("DISPLAY"))

	root := C.XDefaultRootWindow(dpy)
	kc := C.XKeysymToKeycode(dpy, keysym)

	if kc == 0 {
		C.XCloseDisplay(dpy)
		return nil, fmt.Errorf("no keycode for key %q", key)
	}
	log.Printf("hotkey: key=%q -> keycode=%d, modifiers=Alt+Shift", key, kc)

	base := C.uint(C.Mod1Mask | C.ShiftMask)
	mods := []struct {
		mask C.uint
		name string
	}{
		{base, "Alt+Shift"},
		{base | C.Mod2Mask, "Alt+Shift+NumLock"},
		{base | C.LockMask, "Alt+Shift+CapsLock"},
		{base | C.Mod2Mask | C.LockMask, "Alt+Shift+NumLock+CapsLock"},
	}

	grabbed := 0
	for _, m := range mods {
		ret := C.XGrabKey(dpy, C.int(kc), m.mask, root, 1, C.GrabModeAsync, C.GrabModeAsync)
		if ret == 0 {
			log.Printf("hotkey: XGrabKey failed for %s — may be grabbed by another app", m.name)
		} else {
			grabbed++
		}
	}

	if grabbed == 0 {
		C.XCloseDisplay(dpy)
		return nil, fmt.Errorf("failed to grab any hotkey combination")
	}
	log.Printf("hotkey: grabbed %d/%d modifier combos for Alt+Shift+%s", grabbed, len(mods), key)

	m := &Manager{
		display: dpy,
		root:    root,
		showCh:  showCh,
		cancel:  make(chan struct{}),
	}

	go m.loop()
	return m, nil
}

func (m *Manager) loop() {
	log.Printf("hotkey: event loop started, waiting for hotkey")
	for {
		select {
		case <-m.cancel:
			log.Printf("hotkey: event loop cancelled")
			return
		default:
		}

		if C.XPending(m.display) > 0 {
			var ev C.XEvent
			C.XNextEvent(m.display, &ev)
			any := (*C.XAnyEvent)(unsafe.Pointer(&ev))
			if any._type == C.KeyPress {
				log.Printf("hotkey: pressed!")
				select {
				case m.showCh <- true:
				default:
				}
			}
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func (m *Manager) Close() {
	close(m.cancel)
	if m.display != nil {
		// unregister all grabs with the same keycode for cleanup
		// We don't know which keysym was used originally (we don't store it),
		// but XCloseDisplay releases all grabs automatically.
		C.XCloseDisplay(m.display)
		m.display = nil
	}
}

func knownKeys() string {
	keys := make([]string, 0, len(keysymMap))
	for k := range keysymMap {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}
