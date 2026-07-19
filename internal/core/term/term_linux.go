//go:build linux

package term

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	ioctlGetTermios = 0x5401 // TCGETS
	ioctlSetTermios = 0x5402 // TCSETS
	ioctlGetWinsize = 0x5413 // TIOCGWINSZ
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
// ISIG is disabled so Ctrl+C arrives as byte 0x03 (clean cancel, terminal
// always restored). OPOST stays enabled so \n keeps its normal CR+LF
// translation. ICRNL is disabled so Enter arrives as CR (0x0d) and Ctrl+J
// stays distinguishable as LF (0x0a).
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

// InputAvailable reports whether stdin has bytes to read within ms
// milliseconds. Used to disambiguate a lone ESC byte (Esc key) from the
// start of an escape sequence (arrow keys).
func InputAvailable(ms int) bool {
	var fds syscall.FdSet
	fds.Bits[0] = 1
	tv := syscall.NsecToTimeval(int64(ms) * 1e6)
	n, err := syscall.Select(1, &fds, nil, nil, &tv)
	return err == nil && n > 0
}

// ReadInput reads from stdin and disambiguates a lone ESC byte from the
// start of an escape sequence (arrow keys, etc.).
func ReadInput(buf []byte) (int, error) {
	n, err := os.Stdin.Read(buf)
	if err != nil {
		return n, err
	}
	if n == 1 && buf[0] == 0x1b && InputAvailable(25) {
		m, err := os.Stdin.Read(buf[n:])
		if err == nil {
			n += m
		}
	}
	return n, nil
}
