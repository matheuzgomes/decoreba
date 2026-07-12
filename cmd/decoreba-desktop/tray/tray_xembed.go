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
	"image"
	"image/color"
	"image/draw"
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
	dataptr[0] = 0 // CurrentTime
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

func generateIcon() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 24, 24))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 255}}, image.Point{}, draw.Src)

	fg := color.RGBA{255, 255, 255, 255}
	glyph := []struct{ x, y int }{
		{8, 6}, {9, 7}, {10, 8}, {11, 9}, {12, 10}, {13, 11}, {14, 11}, {15, 12},
		{14, 13}, {13, 13}, {12, 14}, {11, 15}, {10, 16}, {9, 17}, {8, 18},
	}
	for _, p := range glyph {
		img.Set(p.x, p.y, fg)
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	buf := make([]byte, 8+w*h*4)
	putLE32(buf[0:4], uint32(w))
	putLE32(buf[4:8], uint32(h))
	off := 8
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixel := uint32(a>>8)<<24 | uint32(r>>8)<<16 | uint32(g>>8)<<8 | uint32(b>>8)
			putLE32(buf[off:off+4], pixel)
			off += 4
		}
	}
	return buf
}

func putLE32(b []byte, v uint32) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}
