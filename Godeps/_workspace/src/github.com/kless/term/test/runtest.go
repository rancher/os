// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package test checks the functions that depend of the standard input,
// which is changed by `go test` to the standard error.
//
// Flags:
//
//  -dbg-key=false: debug: print the decimal code at pressing a key
//  -dbg-winsize=false: debug: to know how many signals are sent at maximizing a window
//  -iact=false: interactive mode
//  -t=2: time in seconds to wait to write in automatic mode
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kless/term"
	"github.com/kless/term/readline"
)

var (
	IsInteractive = flag.Bool("iact", false, "interactive mode")
	Time          = flag.Uint("t", 2, "time in seconds to wait to write in automatic mode")

	DebugWinSize = flag.Bool("dbg-winsize", false, "debug: to know how many signals are sent at maximizing a window")
	DebugKey     = flag.Bool("dbg-key", false, "debug: print the decimal code at pressing a key")

	pr *io.PipeReader
	pw *io.PipeWriter
)

func main() {
	flag.Parse()
	log.SetFlags(0)
	log.SetPrefix("--- FAIL: ")

	if *DebugKey {
		Lookup()
		return
	}
	if *DebugWinSize {
		win := term.DetectWinSize()
		defer win.Close()
		fmt.Println("[Resize the window: should print a number every time]")

		for i := 0; i < 7; i++ {
			select {
			case <-win.Change:
				fmt.Printf("%d ", i)
			case <-time.After(13 * time.Second):
				fmt.Print("\ntimed out\n")
				return
			}
		}
		return
	}

	if !*IsInteractive {
		pr, pw = io.Pipe()
		term.Input = pr
	}

	TestCharMode()
	TestEchoMode()
	if *IsInteractive {
		TestPassword()
	}
	TestEditLine()
	if *IsInteractive {
		TestDetectSize()
	}
}

// TestCharMode tests terminal set to single character mode.
func TestCharMode() {
	fmt.Print("\n=== RUN TestCharMode\n")

	ter, _ := term.New()
	defer func() {
		if err := ter.Restore(); err != nil {
			log.Print(err)
		}
	}()

	if err := ter.CharMode(); err != nil {
		log.Print("expected to set character mode:", err)
		return
	}

	buf := bufio.NewReaderSize(term.Input, 4)
	reply := []string{"a", "€", "~"}

	if !*IsInteractive {
		go func() {
			for _, r := range reply {
				time.Sleep(time.Duration(*Time) * time.Second)
				fmt.Fprint(pw, r)
			}
		}()
	}

	for i := 1; ; i++ {
		fmt.Print(" Press key: ")
		rune, _, err := buf.ReadRune()
		if err != nil {
			log.Print(err)
			return
		}
		fmt.Printf("\n pressed: %q\n", string(rune))

		if *IsInteractive || i == len(reply) {
			break
		}
	}
}

func TestEchoMode() {
	fmt.Print("\n=== RUN TestEchoMode\n")

	ter, _ := term.New()
	defer func() {
		if err := ter.Restore(); err != nil {
			log.Print(err)
		}
	}()

	if err := ter.EchoMode(false); err != nil {
		log.Print("expected to set echo mode:", err)
		return
	}
	fmt.Print(" + Mode to echo off\n")
	buf := bufio.NewReader(term.Input)

	if !*IsInteractive {
		go func() {
			time.Sleep(time.Duration(*Time) * time.Second)
			fmt.Fprint(pw, "Karma\n")
		}()
	}
	fmt.Print(" Write (enter to finish): ")
	line, err := buf.ReadString('\n')
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Printf("\n entered: %q\n", line)

	ter.EchoMode(true)
	fmt.Print("\n + Mode to echo on\n")

	if !*IsInteractive {
		go func() {
			time.Sleep(time.Duration(*Time) * time.Second)
			fmt.Fprint(pw, "hotel\n")
		}()
	}
	fmt.Print(" Write (enter to finish): ")
	line, _ = buf.ReadString('\n')
	if !*IsInteractive {
		fmt.Println()
	}
	fmt.Printf(" entered: %q\n", line)
}

func TestPassword() {
	fmt.Print("\n=== RUN TestPassword\n")

	fmt.Print(" Password (no echo): ")
	pass := make([]byte, 8)

	n, err := term.ReadPassword(pass)
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Printf(" entered: %q\n number: %d\n", pass, n)

	term.PasswordShadowed = true
	fmt.Print("\n Password (shadow character): ")
	pass = make([]byte, 8)

	n, err = term.ReadPassword(pass)
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Printf(" entered: %q\n number: %d\n", pass, n)
}

func TestDetectSize() {
	fmt.Print("\n=== RUN TestDetectSize\n")

	ter, _ := term.New()
	defer func() {
		if err := ter.Restore(); err != nil {
			log.Print(err)
		}
	}()

	row, col, err := ter.GetSize()
	if err != nil {
		panic(err)
	}

	winSize := term.DetectWinSize()
	fmt.Println("[Change the size of the terminal]")

	// I want to finish the test.
	go func() {
		time.Sleep(10 * time.Second)
		winSize.Change <- true
	}()

	<-winSize.Change
	winSize.Close()

	row2, col2, err := ter.GetSize()
	if err != nil {
		panic(err)
	}
	if row == row2 && col == col2 {
		log.Print("the terminal size got the same value")
		return
	}
}

// Package readline

func TestEditLine() {
	fmt.Println("\n=== RUN TestEditLine")

	tempHistory := filepath.Join(os.TempDir(), "test_readline")
	hist, err := readline.NewHistory(tempHistory)
	if err != nil {
		log.Print(err)
		return
	}
	defer func() {
		if err = os.Remove(tempHistory); err != nil {
			log.Print(err)
		}
	}()
	hist.Load()

	fmt.Printf("Press ^D to exit\n\n")

	ln, err := readline.NewDefaultLine(hist)
	if err != nil {
		log.Print(err)
		return
	}
	defer func() {
		if err = ln.Restore(); err != nil {
			log.Print(err)
		}
	}()

	if !*IsInteractive {
		reply := []string{
			"I have heard that the night is all magic",
			"and that a goblin invites you to dream",
		}

		go func() {
			for _, r := range reply {
				time.Sleep(time.Duration(*Time) * time.Second)
				fmt.Fprintf(pw, "%s\r\n", r)
			}
			time.Sleep(time.Duration(*Time) * time.Second)
			pw.Write([]byte{4}) // Ctrl+D
		}()
	}

	for {
		if _, err = ln.Read(); err != nil {
			if err == readline.ErrCtrlD {
				hist.Save()
				err = nil
			} else {
				log.Print(err)
				return
			}
			break
		}
	}
}

// Lookup prints the decimal code at pressing a key.
func Lookup() {
	ter, err := term.New()
	if err != nil {
		log.Print(err)
		return
	}
	defer func() {
		if err = ter.Restore(); err != nil {
			log.Print(err)
		}
	}()

	if err = ter.RawMode(); err != nil {
		log.Print(err)
		return
	} else {
		buf := bufio.NewReader(term.Input)
		runes := make([]int32, 0)
		chars := make([]string, 0)

		fmt.Print("[Press Enter to exit]\r\n")
		fmt.Print("> ")

	L:
		for {
			rune_, _, err := buf.ReadRune()
			if err != nil {
				log.Print(err)
				continue
			}

			switch rune_ {
			default:
				fmt.Print(rune_, " ")
				runes = append(runes, rune_)
				char := strconv.QuoteRune(rune_)
				chars = append(chars, char[1:len(char)-1])
				continue

			case 13:
				fmt.Printf("\r\n\r\n%v\r\n\"%s\"\r\n", runes, strings.Join(chars, " "))
				break L
			}
		}
	}
}
