package tray

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <X11/Xatom.h>
#include <stdlib.h>

#define SYSTEM_TRAY_REQUEST_DOCK 0
*/
import "C"

import (
	"fmt"
	"log"
	"time"
	"unsafe"
)

func xembedProbe() bool {
	cs := C.CString("")
	defer C.free(unsafe.Pointer(cs))
	dpy := C.XOpenDisplay(cs)
	if dpy == nil {
		return false
	}
	defer C.XCloseDisplay(dpy)

	trayAtom := C.XInternAtom(dpy, C.CString("_NET_SYSTEM_TRAY_S0"), C.False)
	owner := C.XGetSelectionOwner(dpy, trayAtom)
	return owner != 0
}

func newXEmbed(showCh chan<- bool, quitCh chan<- struct{}) (*Tray, error) {
	cs := C.CString("")
	defer C.free(unsafe.Pointer(cs))
	dpy := C.XOpenDisplay(cs)
	if dpy == nil {
		return nil, fmt.Errorf("cannot open X display")
	}

	root := C.XDefaultRootWindow(dpy)
	scr := C.XDefaultScreen(dpy)

	trayAtom := C.XInternAtom(dpy, C.CString("_NET_SYSTEM_TRAY_S0"), C.False)
	trayMgr := C.XGetSelectionOwner(dpy, trayAtom)
	if trayMgr == 0 {
		C.XCloseDisplay(dpy)
		return nil, fmt.Errorf("no _NET_SYSTEM_TRAY_S0 manager")
	}

	opcodeAtom := C.XInternAtom(dpy, C.CString("_NET_SYSTEM_TRAY_OPCODE"), C.False)

	var actualType C.Atom
	var actualFormat C.int
	var nItems, bytesAfter C.ulong
	var data *C.uchar

	ret := C.XGetWindowProperty(dpy, trayMgr, opcodeAtom,
		0, 1, C.False, C.XA_ATOM,
		&actualType, &actualFormat, &nItems, &bytesAfter, &data)
	if ret != C.Success || nItems == 0 || data == nil {
		C.XCloseDisplay(dpy)
		return nil, fmt.Errorf("cannot get _NET_SYSTEM_TRAY_OPCODE from manager")
	}

	_ = *(*C.Atom)(unsafe.Pointer(data))
	C.XFree(unsafe.Pointer(data))

	var attr C.XSetWindowAttributes
	attr.background_pixel = C.XBlackPixel(dpy, scr)
	attr.event_mask = C.ExposureMask | C.ButtonPressMask | C.StructureNotifyMask

	win := C.XCreateWindow(dpy, root, 0, 0, 24, 24, 0,
		C.CopyFromParent, C.InputOutput, nil,
		C.CWBackPixel|C.CWEventMask, &attr)

	setIconProp(dpy, win)
	setClassHint(dpy, win)

	var ev C.XClientMessageEvent
	ev._type = C.ClientMessage
	ev.window = trayMgr
	ev.message_type = opcodeAtom
	ev.format = 32
	dataptr := (*[5]C.long)(unsafe.Pointer(&ev.data[0]))
	dataptr[0] = 0
	dataptr[1] = C.long(C.SYSTEM_TRAY_REQUEST_DOCK)
	dataptr[2] = C.long(win)

	C.XSendEvent(dpy, trayMgr, C.False, C.NoEventMask, (*C.XEvent)(unsafe.Pointer(&ev)))
	C.XFlush(dpy)

	log.Printf("tray XEmbed: docked window 0x%x to manager 0x%x", uint64(win), uint64(trayMgr))

	cancel := make(chan struct{})

	go func() {
		for {
			select {
			case <-cancel:
				return
			default:
			}

			if C.XPending(dpy) > 0 {
				var xev C.XEvent
				C.XNextEvent(dpy, &xev)
				any := (*C.XAnyEvent)(unsafe.Pointer(&xev))
				if any._type == C.ButtonPress {
					log.Printf("tray XEmbed: click detected")
					select {
					case showCh <- true:
					default:
					}
				} else if any._type == C.Expose {
					redrawIcon(dpy, win, scr)
				}
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()

	return &Tray{closeFn: func() error {
		close(cancel)
		time.Sleep(150 * time.Millisecond)
		if win != 0 {
			C.XDestroyWindow(dpy, win)
		}
		C.XCloseDisplay(dpy)
		return nil
	}}, nil
}

func setIconProp(dpy *C.Display, win C.Window) {
	img := generateIcon()
	netWmIcon := C.XInternAtom(dpy, C.CString("_NET_WM_ICON"), C.False)
	if netWmIcon == 0 {
		return
	}
	C.XChangeProperty(dpy, win, netWmIcon, C.XA_CARDINAL, 32,
		C.PropModeReplace, (*C.uchar)(unsafe.Pointer(&img[0])), C.int(len(img)/4))
	C.XFlush(dpy)
}

func setClassHint(dpy *C.Display, win C.Window) {
	name := C.CString("decoreba")
	class := C.CString("Decoreba")
	defer C.free(unsafe.Pointer(name))
	defer C.free(unsafe.Pointer(class))

	var hint C.XClassHint
	hint.res_name = name
	hint.res_class = class
	C.XSetClassHint(dpy, win, &hint)
}

func redrawIcon(dpy *C.Display, win C.Window, scr C.int) {
	gc := C.XCreateGC(dpy, win, 0, nil)
	bg := C.XBlackPixel(dpy, scr)
	fg := C.XWhitePixel(dpy, scr)

	C.XSetForeground(dpy, gc, bg)
	C.XFillRectangle(dpy, win, gc, 0, 0, 24, 24)

	C.XSetForeground(dpy, gc, fg)
	points := []C.XPoint{
		{8, 6},
		{16, 12},
		{8, 18},
	}
	C.XDrawLines(dpy, win, gc, &points[0], 3, C.CoordModeOrigin)

	C.XFreeGC(dpy, gc)
	C.XFlush(dpy)
}
