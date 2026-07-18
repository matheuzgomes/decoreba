//go:build windows

package term

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode             = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode             = kernel32.NewProc("SetConsoleMode")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	procGetConsoleCP               = kernel32.NewProc("GetConsoleCP")
	procSetConsoleCP               = kernel32.NewProc("SetConsoleCP")
	procGetConsoleOutputCP         = kernel32.NewProc("GetConsoleOutputCP")
	procSetConsoleOutputCP         = kernel32.NewProc("SetConsoleOutputCP")
)

const (
	enableProcessedInput = 0x0001
	enableLineInput      = 0x0002
	enableEchoInput      = 0x0004
	enableVTInput        = 0x0200
	enableVTProcessing   = 0x0004
	cpUTF8               = 65001
)

func getConsoleMode(fd uintptr) (uint32, error) {
	var mode uint32
	r, _, err := procGetConsoleMode.Call(fd, uintptr(unsafe.Pointer(&mode)))
	if r == 0 {
		return 0, err
	}
	return mode, nil
}

func setConsoleMode(fd uintptr, mode uint32) error {
	r, _, err := procSetConsoleMode.Call(fd, uintptr(mode))
	if r == 0 {
		return err
	}
	return nil
}

func IsTerminal() bool {
	_, err := getConsoleMode(os.Stdin.Fd())
	return err == nil
}

// MakeRaw switches the console to raw + VT mode and returns a restore function.
func MakeRaw() (func(), error) {
	inFd := os.Stdin.Fd()
	outFd := os.Stdout.Fd()

	oldIn, err := getConsoleMode(inFd)
	if err != nil {
		return nil, err
	}
	oldOut, outErr := getConsoleMode(outFd)
	oldInCP, _, _ := procGetConsoleCP.Call()
	oldOutCP, _, _ := procGetConsoleOutputCP.Call()

	newIn := (oldIn &^ (enableLineInput | enableEchoInput | enableProcessedInput)) | enableVTInput
	if err := setConsoleMode(inFd, newIn); err != nil {
		return nil, err
	}
	if outErr == nil {
		_ = setConsoleMode(outFd, oldOut|enableVTProcessing)
	}
	procSetConsoleCP.Call(cpUTF8)
	procSetConsoleOutputCP.Call(cpUTF8)

	return func() {
		_ = setConsoleMode(inFd, oldIn)
		if outErr == nil {
			_ = setConsoleMode(outFd, oldOut)
		}
		procSetConsoleCP.Call(oldInCP)
		procSetConsoleOutputCP.Call(oldOutCP)
	}, nil
}

type consoleScreenBufferInfo struct {
	Size, CursorPosition [2]int16
	Attributes           uint16
	Window               [4]int16
	MaximumWindowSize    [2]int16
}

func GetSize() (width, height int) {
	var info consoleScreenBufferInfo
	r, _, _ := procGetConsoleScreenBufferInfo.Call(os.Stdout.Fd(), uintptr(unsafe.Pointer(&info)))
	if r == 0 {
		return 0, 0
	}
	return int(info.Window[2] - info.Window[0] + 1), int(info.Window[3] - info.Window[1] + 1)
}

func InputAvailable(ms int) bool {
	return false
}
