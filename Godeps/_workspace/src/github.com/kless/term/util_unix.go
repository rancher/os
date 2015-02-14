// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// +build !plan9,!windows

package term

//#include <unistd.h>
//import "C"

import (
	"bytes"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/kless/term/sys"
)

var shellsWithoutANSI = []string{"dumb", "cons25"}

// SupportANSI checks if the terminal supports ANSI escape sequences.
func SupportANSI() bool {
	term := os.Getenv("TERM")
	if term == "" {
		return false
	}

	for _, v := range shellsWithoutANSI {
		if v == term {
			return false
		}
	}
	return true
}

// char *ttyname(int fd)
// http://sourceware.org/git/?p=glibc.git;a=blob;f=sysdeps/unix/sysv/linux/ttyname.c;hb=HEAD
// http://sourceware.org/git/?p=glibc.git;a=blob;f=sysdeps/posix/ttyname.c;hb=HEAD
// GetName gets the name of a term.
/*func GetName(fd int) (string, error) {
	name, errno := C.ttyname(C.int(fd))
	if errno != nil {
		return "", fmt.Errorf("term.TTYName: %s", errno)
	}
	return C.GoString(name), nil
}*/

// int isatty(int fd)
// http://sourceware.org/git/?p=glibc.git;a=blob;f=sysdeps/posix/isatty.c;hb=HEAD

// IsTerminal returns true if the file descriptor is a term.
func IsTerminal(fd int) bool {
	return sys.Getattr(fd, &sys.Termios{}) == nil
}

// ReadPassword reads characters from the input until press Enter or until
// fill in the given slice.
//
// Only reads characters that include letters, marks, numbers, punctuation,
// and symbols from Unicode categories L, M, N, P, S, besides of the
// ASCII space character.
// Ctrl-C interrumpts, and backspace removes the last character read.
//
// Returns the number of bytes read.
func ReadPassword(password []byte) (n int, err error) {
	ter, err := New()
	if err != nil {
		return 0, err
	}
	defer func() {
		err2 := ter.Restore()
		if err2 != nil && err == nil {
			err = err2
		}
	}()

	if err = ter.RawMode(); err != nil {
		return 0, err
	}

	key := make([]byte, 4) // In-memory representation of a rune.
	lenPassword := 0       // Number of characters read.

	if PasswordShadowed {
		rand.Seed(int64(time.Now().Nanosecond()))
	}

L:
	for {
		n, err = syscall.Read(InputFD, key)
		if err != nil {
			return 0, err
		}

		if n == 1 {
			switch key[0] {
			case sys.K_RETURN:
				break L
			case sys.K_BACK:
				if lenPassword != 0 {
					lenPassword--
					password[lenPassword] = 0
				}
				continue
			case sys.K_CTRL_C:
				syscall.Write(syscall.Stdout, _CTRL_C)
				// Clean data stored, if any.
				for i, v := range password {
					if v == 0 {
						break
					}
					password[i] = 0
				}
				return 0, nil
			}
		}

		char, _ := utf8.DecodeRune(key)
		if unicode.IsPrint(char) {
			password[lenPassword] = key[0] // Only want a character by key
			lenPassword++

			if PasswordShadowed {
				syscall.Write(syscall.Stdout, bytes.Repeat(_SHADOW_CHAR, rand.Intn(3)+1))
			}
			if lenPassword == len(password) {
				break
			}
		}
	}

	syscall.Write(syscall.Stdout, _RETURN)
	n = lenPassword
	return
}

// WinSize represents a channel, Change, to know when the window size has
// changed through function DetectWinSize.
type WinSize struct {
	Change chan bool
	quit   chan bool
	wait   chan bool
}

// DetectWinSize caughts a signal named SIGWINCH whenever the window size changes,
// being indicated in channel `WinSize.Change`.
func DetectWinSize() *WinSize {
	w := &WinSize{
		make(chan bool),
		make(chan bool),
		make(chan bool),
	}

	changeSig := make(chan os.Signal)
	signal.Notify(changeSig, syscall.SIGWINCH)

	go func() {
		for {
			select {
			case <-changeSig:
				// Add a pause because it is sent two signals at maximizing a window.
				time.Sleep(7 * time.Millisecond)
				w.Change <- true
			case <-w.quit:
				w.wait <- true
				return
			}
		}
	}()
	return w
}

// Close closes the goroutine started to trap the signal.
func (w *WinSize) Close() {
	w.quit <- true
	<-w.wait
}
