//go:build darwin

package term

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	ioctlGetTermios = 0x40487413 // TIOCGETA
	ioctlSetTermios = 0x80487414 // TIOCSETA
	ioctlGetWinsize = 0x40087468 // TIOCGWINSZ
)

func getTermios(fd uintptr) (*syscall.Termios, error) {
	var t syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, ioctlGetTermios, uintptr(unsafe.Pointer(&t)))
	if errno != 0 {
		return nil, errno
	}
	return &t, nil
}

func setTermios(fd uintptr, t *syscall.Termios) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, ioctlSetTermios, uintptr(unsafe.Pointer(t)))
	if errno != 0 {
		return errno
	}
	return nil
}

func IsTerminal() bool {
	_, err := getTermios(os.Stdin.Fd())
	return err == nil
}

// MakeRaw puts stdin in raw mode and returns a restore function.
func MakeRaw() (func(), error) {
	fd := os.Stdin.Fd()
	old, err := getTermios(fd)
	if err != nil {
		return nil, err
	}
	raw := *old
	raw.Iflag &^= syscall.IXON | syscall.ICRNL
	raw.Lflag &^= syscall.ICANON | syscall.ECHO | syscall.IEXTEN | syscall.ISIG
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	if err := setTermios(fd, &raw); err != nil {
		return nil, err
	}
	return func() { _ = setTermios(fd, old) }, nil
}

func GetSize() (width, height int) {
	var ws struct {
		Row, Col, Xpixel, Ypixel uint16
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(), ioctlGetWinsize, uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return 0, 0
	}
	return int(ws.Col), int(ws.Row)
}

func InputAvailable(ms int) bool {
	var fds syscall.FdSet
	fds.Bits[0] = 1
	tv := syscall.NsecToTimeval(int64(ms) * 1e6)
	if err := syscall.Select(1, &fds, nil, nil, &tv); err != nil {
		return false
	}
	return fds.Bits[0]&1 != 0
}
