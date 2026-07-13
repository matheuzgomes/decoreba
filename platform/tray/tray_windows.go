package tray

/*
#cgo LDFLAGS: -luser32 -lshell32
#include <windows.h>
#include <shellapi.h>
#include <stdlib.h>

#define WM_TRAY_CALLBACK (WM_USER + 1)
#define ID_TRAY_ICON 1

extern void goTrayClickC();

static LRESULT CALLBACK trayWndProc(HWND hwnd, UINT msg, WPARAM wp, LPARAM lp) {
	if (msg == WM_TRAY_CALLBACK && lp == WM_LBUTTONDOWN) {
		goTrayClickC();
	}
	return DefWindowProc(hwnd, msg, wp, lp);
}

static int registerTrayClass(HINSTANCE hInstance) {
	WNDCLASS wc;
	memset(&wc, 0, sizeof(wc));
	wc.lpfnWndProc = trayWndProc;
	wc.hInstance = hInstance;
	wc.lpszClassName = "DecorebaTrayClass";
	wc.hIcon = LoadIcon(NULL, IDI_APPLICATION);
	return RegisterClass(&wc) != 0;
}

static HWND createTrayWindow(HINSTANCE hInstance) {
	return CreateWindowEx(0, "DecorebaTrayClass", "", 0, 0, 0, 0, 0, NULL, NULL, hInstance, NULL);
}

static int addTrayIcon(HWND hwnd, HINSTANCE hInst) {
	NOTIFYICONDATA nid;
	memset(&nid, 0, sizeof(nid));
	nid.cbSize = sizeof(nid);
	nid.hWnd = hwnd;
	nid.uID = ID_TRAY_ICON;
	nid.uFlags = NIF_ICON | NIF_MESSAGE | NIF_TIP;
	nid.uCallbackMessage = WM_TRAY_CALLBACK;
	nid.hIcon = LoadIcon(hInst, MAKEINTRESOURCE(1));
	if (!nid.hIcon) nid.hIcon = LoadIcon(NULL, IDI_APPLICATION);
	strncpy(nid.szTip, "decoreba", sizeof(nid.szTip));
	return Shell_NotifyIcon(NIM_ADD, &nid);
}

static int removeTrayIcon(HWND hwnd) {
	NOTIFYICONDATA nid;
	memset(&nid, 0, sizeof(nid));
	nid.cbSize = sizeof(nid);
	nid.hWnd = hwnd;
	nid.uID = ID_TRAY_ICON;
	return Shell_NotifyIcon(NIM_DELETE, &nid);
}
*/
import "C"
import (
	"log"
	"runtime"
)

func probePlatform() bool {
	return true
}

var (
	trayShowCh chan<- bool
	trayQuitCh chan<- struct{}
)

func newPlatform(showCh chan<- bool, quitCh chan<- struct{}) (*Tray, error) {
	trayShowCh = showCh
	trayQuitCh = quitCh

	cancel := make(chan struct{})

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		hInstance := C.GetModuleHandle(nil)
		if C.registerTrayClass(hInstance) == 0 {
			log.Printf("tray: RegisterClass failed")
			return
		}

		hwnd := C.createTrayWindow(hInstance)
		if hwnd == nil {
			log.Printf("tray: CreateWindowEx failed")
			return
		}

		if C.addTrayIcon(hwnd, hInstance) == 0 {
			log.Printf("tray: Shell_NotifyIcon(NIM_ADD) failed")
			C.DestroyWindow(hwnd)
			return
		}

		log.Printf("tray: started (Windows)")

		for {
			select {
			case <-cancel:
				C.removeTrayIcon(hwnd)
				C.DestroyWindow(hwnd)
				return
			default:
			}

			var msg C.MSG
			if C.PeekMessage(&msg, nil, 0, 0, C.PM_REMOVE) != 0 {
				C.TranslateMessage(&msg)
				C.DispatchMessage(&msg)
			} else {
				C.Sleep(50)
			}
		}
	}()

	return &Tray{closeFn: func() error {
		close(cancel)
		return nil
	}}, nil
}

//export goTrayClickC
func goTrayClickC() {
	select {
	case trayShowCh <- true:
	default:
	}
}
