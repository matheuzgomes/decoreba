package tray

/*
#cgo LDFLAGS: -framework AppKit
#include <stdlib.h>
#include <objc/runtime.h>
#include <objc/message.h>

extern void trayActivateC();
extern void trayQuitC();

static void* statusItem = NULL;

static void trayInit() {
	id cls = (id)objc_getClass("NSStatusBar");
	id bar = ((id (*)(id, SEL))objc_msgSend)(cls, sel_getUid("systemStatusBar"));
	statusItem = ((id (*)(id, SEL, double))objc_msgSend)(bar, sel_getUid("statusItemWithLength:"), -2.0);

	id button = ((id (*)(id, SEL))objc_msgSend)(statusItem, sel_getUid("button"));
	id title = ((id (*)(id, SEL, const char*))objc_msgSend)((id)objc_getClass("NSString"), sel_getUid("stringWithUTF8String:"), ">_");
	((void (*)(id, SEL, id))objc_msgSend)(button, sel_getUid("setTitle:"), title);

	id target = ((id (*)(id, SEL))objc_msgSend)((id)objc_getClass("NSObject"), sel_getUid("alloc"));
	target = ((id (*)(id, SEL))objc_msgSend)(target, sel_getUid("init"));

	class_addMethod((Class)objc_getClass("NSObject"), sel_getUid("trayAction:"), (IMP)trayActivateC, "v@:@");

	((void (*)(id, SEL, id))objc_msgSend)(button, sel_getUid("setTarget:"), target);
	((void (*)(id, SEL, SEL))objc_msgSend)(button, sel_getUid("setAction:"), sel_getUid("trayAction:"));
}

static void trayCleanup() {
	if (statusItem) {
		id cls = (id)objc_getClass("NSStatusBar");
		id bar = ((id (*)(id, SEL))objc_msgSend)(cls, sel_getUid("systemStatusBar"));
		((void (*)(id, SEL, id))objc_msgSend)(bar, sel_getUid("removeStatusItem:"), statusItem);
		statusItem = NULL;
	}
}
*/
import "C"
import (
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

	runtime.LockOSThread()

	C.trayInit()

	return &Tray{closeFn: func() error {
		C.trayCleanup()
		runtime.UnlockOSThread()
		return nil
	}}, nil
}

//export trayActivateC
func trayActivateC() {
	select {
	case trayShowCh <- true:
	default:
	}
}

//export trayQuitC
func trayQuitC() {
	select {
	case trayQuitCh <- struct{}{}:
	default:
	}
}
