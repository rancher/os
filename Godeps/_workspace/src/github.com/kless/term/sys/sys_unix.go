// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// +build !plan9,!windows

// Reference: man termios ; man tty_ioctl
// Linux file: "asm-generic/termbits.h"

package sys

import (
	"syscall"
	"unsafe"
)

// int tcgetattr(int fd, struct termios *termios_p)

func Getattr(fd int, state *Termios) (err error) {
	_, _, e1 := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(TCGETS), uintptr(unsafe.Pointer(state)))
	if e1 != 0 {
		err = e1
	}
	return
}

// int tcsetattr(int fd, int optional_actions, const struct termios *termios_p)

func Setattr(fd int, action uint, state *Termios) (err error) {
	switch action {
	case TCSANOW:
		action = TCSETS
	case TCSADRAIN:
		action = TCSETSW
	case TCSAFLUSH:
		action = TCSETSF
	}

	_, _, e1 := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(action), uintptr(unsafe.Pointer(state)))
	if e1 != 0 {
		err = e1
	}
	return
}

// GetWinsize gets the winsize struct with the terminal size set by the kernel.
func GetWinsize(fd int, ws *Winsize) (err error) {
	_, _, e1 := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(TIOCGWINSZ), uintptr(unsafe.Pointer(ws)))
	if e1 != 0 {
		err = e1
	}
	return
}

// Types

//cgo const (TCSANOW, TCSADRAIN, TCSAFLUSH)
//cgo const TIOCGWINSZ

//cgo type struct_termios
//cgo type struct_winsize

//cgo// c_cc characters
//cgo const (VINTR, VQUIT, VERASE, VKILL, VEOF, VTIME, VMIN, VSTART, VSTOP,
// VSUSP, VEOL, VREPRINT, VDISCARD, VWERASE, VLNEXT, VEOL2)

//cgo// c_iflag bits
//cgo const (IGNBRK, BRKINT, IGNPAR, PARMRK, INPCK, ISTRIP, INLCR, IGNCR, ICRNL,
// IXON, IXANY, IXOFF, IMAXBEL)

//cgo// c_oflag bits
//cgo const (OPOST, ONLCR, OCRNL, ONOCR, ONLRET, NL0, NL1, CR0, CR1, CR2, CR3,
// TAB0, TAB1, TAB2, XTABS, BS0, BS1, FF0, FF1)

//cgo// c_cflag bits
//cgo const (B0, B50, B75, B110, B134, B150, B200, B300,
// B600, B1200, B1800, B2400, B4800, B9600, B19200, B38400, EXTA, EXTB, CSIZE,
// CS5, CS6, CS7, CS8, CSTOPB, CREAD, PARENB, PARODD, HUPCL, CLOCAL,
// B57600, B115200, B230400, CRTSCTS)

//cgo// c_lflag bits
//cgo const (ISIG, ICANON, ECHO, ECHOE, ECHOK, ECHONL, NOFLSH, TOSTOP, ECHOCTL,
// ECHOPRT, ECHOKE, FLUSHO, PENDIN, IEXTEN, EXTPROC)

/*
== FreeBSD has not:

// c_cc characters
VSWTC

// c_iflag bits
(IUCLC, IUTF8)

// c_oflag bits
(OLCUC, OFILL, OFDEL, NLDLY, CRDLY, BSDLY, VTDLY, VT0, VT1, FFDLY)

// c_cflag bits
(CBAUD, CBAUDEX, BOTHER, B500000, B576000,
B1000000, B1152000, B1500000, B2000000, B2500000, B3000000, B3500000,
B4000000, CIBAUD, CMSPAR, IBSHIFT)

// c_lflag bits
XCASE

== NetBSD, besides, has not:

// c_oflag bits
(TABDLY, TAB3)

== OpenBSD, besides, has not:

// c_cflag bits
(B460800, B921600)
*/
