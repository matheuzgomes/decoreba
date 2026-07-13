package tray

/*
#cgo LDFLAGS: -luser32 -lshell32
#include <windows.h>
#include <shellapi.h>
#include <stdlib.h>

#define WM_TRAY_CALLBACK (WM_USER + 1)
#define ID_TRAY_ICON 1
*/
import "C"
import (
	"fmt"
	"log"
	"runtime"
	"time"
	"unsafe"
)

func probePlatform() bool {
	return true
}

func newPlatform(showCh chan<- bool, quitCh chan<- struct{}) (*Tray, error) {
	runtime.LockOSThread()

	name := C.CString("STATIC")
	defer C.free(unsafe.Pointer(name))

	hwnd := C.CreateWindowEx(0, name, nil, 0, 0, 0, 0, 0, nil, nil,
		C.GetModuleHandle(nil), nil)
	if hwnd == nil {
		runtime.UnlockOSThread()
		return nil, fmt.Errorf("CreateWindowEx failed")
	}

	icon := createHIcon()

	var nid C.NOTIFYICONDATA
	C.memset(unsafe.Pointer(&nid), 0, C.sizeof_NOTIFYICONDATA)
	nid.cbSize = C.sizeof_NOTIFYICONDATA
	nid.hWnd = hwnd
	nid.uID = C.ID_TRAY_ICON
	nid.uFlags = C.NIF_ICON | C.NIF_MESSAGE | C.NIF_TIP
	nid.uCallbackMessage = C.WM_TRAY_CALLBACK
	nid.hIcon = icon

	tip := "decoreba"
	for i := 0; i < len(tip) && i < 128; i++ {
		nid.szTip[i] = C.CHAR(tip[i])
	}

	ret := C.Shell_NotifyIcon(C.NIM_ADD, &nid)
	if ret == 0 {
		if icon != nil {
			C.DestroyIcon(icon)
		}
		C.DestroyWindow(hwnd)
		runtime.UnlockOSThread()
		return nil, fmt.Errorf("Shell_NotifyIcon(NIM_ADD) failed")
	}

	log.Printf("tray: Shell_NotifyIcon success")

	cancel := make(chan struct{})

	go func() {
		defer runtime.UnlockOSThread()
		defer C.DestroyWindow(hwnd)

		for {
			select {
			case <-cancel:
				return
			default:
			}

			var msg C.MSG
			ret := C.PeekMessage(&msg, nil, 0, 0, C.PM_REMOVE)
			if ret != 0 {
				if msg.message == C.WM_TRAY_CALLBACK && C.LPARAM(msg.lParam) == C.LPARAM(C.WM_LBUTTONDOWN) {
					log.Printf("tray: click detected")
					select {
					case showCh <- true:
					default:
					}
				}
				C.TranslateMessage(&msg)
				C.DispatchMessage(&msg)
			} else {
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	return &Tray{closeFn: func() error {
		close(cancel)
		time.Sleep(150 * time.Millisecond)
		C.Shell_NotifyIcon(C.NIM_DELETE, &nid)
		if icon != nil {
			C.DestroyIcon(icon)
		}
		return nil
	}}, nil
}

func createHIcon() C.HICON {
	img := generateIcon()
	if len(img) < 8 {
		return nil
	}

	w := int32(img[0]) | int32(img[1])<<8 | int32(img[2])<<16 | int32(img[3])<<24
	h := int32(img[4]) | int32(img[5])<<8 | int32(img[6])<<16 | int32(img[7])<<24

	hdc := C.GetDC(nil)
	defer C.ReleaseDC(nil, hdc)

	var bmi C.BITMAPINFO
	bmi.bmiHeader.biSize = C.DWORD(C.sizeof_BITMAPINFOHEADER)
	bmi.bmiHeader.biWidth = C.LONG(w)
	bmi.bmiHeader.biHeight = C.LONG(-h)
	bmi.bmiHeader.biPlanes = 1
	bmi.bmiHeader.biBitCount = 32
	bmi.bmiHeader.biCompression = C.BI_RGB

	var pBits unsafe.Pointer
	hColor := C.CreateDIBSection(hdc, &bmi, C.DIB_RGB_COLORS, &pBits, nil, 0)
	if hColor == nil {
		return nil
	}
	defer C.DeleteObject(C.HGDIOBJ(hColor))

	var maskBmi C.BITMAPINFO
	maskBmi.bmiHeader.biSize = C.DWORD(C.sizeof_BITMAPINFOHEADER)
	maskBmi.bmiHeader.biWidth = C.LONG(w)
	maskBmi.bmiHeader.biHeight = C.LONG(-h)
	maskBmi.bmiHeader.biPlanes = 1
	maskBmi.bmiHeader.biBitCount = 1
	maskBmi.bmiHeader.biCompression = C.BI_RGB

	var pMaskBits unsafe.Pointer
	hMask := C.CreateDIBSection(hdc, &maskBmi, C.DIB_RGB_COLORS, &pMaskBits, nil, 0)
	if hMask == nil {
		return nil
	}
	defer C.DeleteObject(C.HGDIOBJ(hMask))

	if pBits != nil {
		pixelCount := int(w) * int(h)
		dst := unsafe.Slice((*byte)(pBits), pixelCount*4)
		src := img[8:]
		for i := 0; i < pixelCount*4 && i < len(src); i++ {
			dst[i] = src[i]
		}
	}

	if pMaskBits != nil {
		maskSize := ((int(w) + 15) / 16 * 2) * int(h)
		for i := 0; i < maskSize; i++ {
			*(*byte)(unsafe.Add(pMaskBits, i)) = 0xFF
		}
	}

	var ii C.ICONINFO
	ii.fIcon = C.TRUE
	ii.hbmColor = C.HBITMAP(hColor)
	ii.hbmMask = C.HBITMAP(hMask)

	return C.CreateIconIndirect(&ii)
}
