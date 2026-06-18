//go:build linux

package repl

import (
	"syscall"
	"unsafe"
)

func isTerminal(fd int) bool {
	_, err := ioctlGetTermios(fd, syscall.TCGETS)
	return err == nil
}

func makeRaw(fd int) (*syscall.Termios, error) {
	oldState, err := ioctlGetTermios(fd, syscall.TCGETS)
	if err != nil {
		return nil, err
	}

	raw := *oldState
	raw.Iflag &^= syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP | syscall.IXON
	raw.Oflag |= syscall.OPOST | syscall.ONLCR
	raw.Cflag |= syscall.CS8
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0

	if err := ioctlSetTermios(fd, syscall.TCSETS, &raw); err != nil {
		return nil, err
	}
	return oldState, nil
}

func restoreTerminal(fd int, state *syscall.Termios) error {
	if state == nil {
		return nil
	}
	return ioctlSetTermios(fd, syscall.TCSETS, state)
}

func ioctlGetTermios(fd int, request uintptr) (*syscall.Termios, error) {
	var state syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), request, uintptr(unsafe.Pointer(&state)))
	if errno != 0 {
		return nil, errno
	}
	return &state, nil
}

func ioctlSetTermios(fd int, request uintptr, state *syscall.Termios) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), request, uintptr(unsafe.Pointer(state)))
	if errno != 0 {
		return errno
	}
	return nil
}
