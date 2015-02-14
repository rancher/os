// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// +build !plan9,!windows

package readline

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/kless/term"
	"github.com/kless/term/sys"
)

func init() {
	if !term.SupportANSI() {
		panic("Your terminal does not support ANSI")
	}
}

// NewLine returns a line using both prompts ps1 and ps2, and setting the given
// terminal to raw mode, if were necessary.
// lenAnsi is the length of ANSI codes that the prompt ps1 could have.
// If the history is nil then it is not used.
func NewLine(ter *term.Terminal, ps1, ps2 string, lenAnsi int, hist *history) (*Line, error) {
	if ter.Mode()&term.RawMode == 0 { // the raw mode is not set
		if err := ter.RawMode(); err != nil {
			return nil, err
		}
	}

	lenPS1 := len(ps1) - lenAnsi
	_, col, err := ter.GetSize()
	if err != nil {
		return nil, err
	}

	buf := newBuffer(lenPS1, col)
	buf.insertRunes([]rune(ps1))

	return &Line{
		ter: ter,
		buf: buf,
		hist: hist,

		ps1: ps1,
		ps2: ps2,
		lenPS1: lenPS1,

		useHistory: hasHistory(hist),
	}, nil
}

// Prompt prints the primary prompt.
func (ln *Line) Prompt() (err error) {
	if _, err = term.Output.Write(DelLine_CR); err != nil {
		return outputError(err.Error())
	}
	if _, err = fmt.Fprint(term.Output, ln.ps1); err != nil {
		return outputError(err.Error())
	}

	ln.buf.pos, ln.buf.size = ln.lenPS1, ln.lenPS1
	return
}

// Read reads charactes from input to write them to output, enabling line editing.
// The errors that could return are to indicate if Ctrl+D was pressed, and for
// both input/output errors.
func (ln *Line) Read() (line string, err error) {
	var anotherLine []rune // For lines got from history.
	var isHistoryUsed bool // If the history has been accessed.
	var action keyAction

	in := bufio.NewReader(term.Input) // Read input.
	esc := make([]byte, 2)            // For escape sequences.
	extEsc := make([]byte, 3)         // Extended escape sequences.

	// Print the primary prompt.
	if err = ln.Prompt(); err != nil {
		return "", err
	}

	// == Detect change of window size.
	winSize := term.DetectWinSize()

	go func() {
		for {
			select {
			case <-winSize.Change: // Wait for.
				_, col, err := ln.ter.GetSize()
				if err != nil {
					ln.buf.columns = col
					ln.buf.refresh()
				}
			}
		}
	}()
	defer winSize.Close()

	for ; ; action = 0 {
		char, _, err := in.ReadRune()
		if err != nil {
			return "", inputError(err.Error())
		}

	_S:
		switch char {
		default:
			if err = ln.buf.insertRune(char); err != nil {
				return "", err
			}
			continue

		case sys.K_RETURN:
			line = ln.buf.toString()

			if ln.useHistory {
				ln.hist.Add(line)
			}
			if _, err = term.Output.Write(CRLF); err != nil {
				return "", outputError(err.Error())
			}
			return strings.TrimSpace(line), nil

		case sys.K_TAB:
			// TODO: disabled by now
			continue

		case sys.K_BACK, sys.K_CTRL_H:
			if err = ln.buf.deleteCharPrev(); err != nil {
				return "", err
			}
			continue

		case sys.K_CTRL_C:
			if err = ln.buf.insertRunes(CtrlC); err != nil {
				return "", err
			}
			if _, err = term.Output.Write(CRLF); err != nil {
				return "", outputError(err.Error())
			}

			ChanCtrlC <- 1 //TODO: is really necessary?

			if err = ln.Prompt(); err != nil {
				return "", err
			}
			continue
		case sys.K_CTRL_D:
			if err = ln.buf.insertRunes(CtrlD); err != nil {
				return "", err
			}
			if _, err = term.Output.Write(CRLF); err != nil {
				return "", outputError(err.Error())
			}

			ln.Restore()
			ChanCtrlD <- 1
			return "", ErrCtrlD

		// Escape sequence
		case sys.K_ESCAPE: // Ctrl+[ ("\x1b" in hexadecimal, "033" in octal)
			if _, err = in.Read(esc); err != nil {
				return "", inputError(err.Error())
			}

			if esc[0] == 79 { // 'O'
				switch esc[1] {
				case 72: // Home: "\x1b O H"
					action = _HOME
					break _S
				case 70: // End: "\x1b O F"
					action = _END
					break _S
				}
			}

			if esc[0] == 91 { // '['
				switch esc[1] {
				case 65: // Up: "\x1b [ A"
					if !ln.useHistory {
						continue
					}
					action = _UP
					break _S
				case 66: // Down: "\x1b [ B"
					if !ln.useHistory {
						continue
					}
					action = _DOWN
					break _S
				case 68: // "\x1b [ D"
					action = _LEFT
					break _S
				case 67: // "\x1b [ C"
					action = _RIGHT
					break _S
				}

				// Extended escape.
				if esc[1] > 48 && esc[1] < 55 {
					if _, err = in.Read(extEsc); err != nil {
						return "", inputError(err.Error())
					}

					if extEsc[0] == 126 { // '~'
						switch esc[1] {
						//case 50: // Insert: "\x1b [ 2 ~"

						case 51: // Delete: "\x1b [ 3 ~"
							if err = ln.buf.deleteChar(); err != nil {
								return "", err
							}
							continue
							//case 53: // RePag: "\x1b [ 5 ~"

							//case 54: // AvPag: "\x1b [ 6 ~"

						}
					}
					if esc[1] == 49 && extEsc[0] == 59 && extEsc[1] == 53 { // "1;5"
						switch extEsc[2] {
						case 68: // Ctrl+left arrow: "\x1b [ 1 ; 5 D"
							// move to last word
							if err = ln.buf.wordBackward(); err != nil {
								return "", err
							}
							continue
						case 67: // Ctrl+right arrow: "\x1b [ 1 ; 5 C"
							// move to next word
							if err = ln.buf.wordForward(); err != nil {
								return "", err
							}
							continue
						}
					}
				}
			}
			continue

		case sys.K_CTRL_T: // Swap actual character by the previous one.
			if err = ln.buf.swap(); err != nil {
				return "", err
			}
			continue

		case sys.K_CTRL_L: // Clear screen.
			if _, err = term.Output.Write(DelScreenToUpper); err != nil {
				return "", err
			}
			if err = ln.Prompt(); err != nil {
				return "", err
			}
			continue
		case sys.K_CTRL_U: // Delete the whole line.
			if err = ln.buf.deleteLine(); err != nil {
				return "", err
			}
			if err = ln.Prompt(); err != nil {
				return "", err
			}
			continue
		case sys.K_CTRL_K: // Delete from current to end of line.
			if err = ln.buf.deleteToRight(); err != nil {
				return "", err
			}
			continue

		case sys.K_CTRL_P: // Up
			if !ln.useHistory {
				continue
			}
			action = _UP
		case sys.K_CTRL_N: // Down
			if !ln.useHistory {
				continue
			}
			action = _DOWN
		case sys.K_CTRL_B: // Left
			action = _LEFT
		case sys.K_CTRL_F: // Right
			action = _RIGHT

		case sys.K_CTRL_A: // Start of line.
			action = _HOME
		case sys.K_CTRL_E: // End of line.
			action = _END
		}

		switch action {
		case _UP, _DOWN: // Up and down arrow: history
			if action == _UP {
				anotherLine, err = ln.hist.Prev()
			} else {
				anotherLine, err = ln.hist.Next()
			}
			if err != nil {
				continue
			}

			// Update the current history entry before to overwrite it with
			// the next one.
			// TODO: it has to be removed before of to be saved the history
			if !isHistoryUsed {
				ln.hist.Add(ln.buf.toString())
			}
			isHistoryUsed = true

			ln.buf.grow(len(anotherLine))
			ln.buf.size = len(anotherLine) + ln.buf.promptLen
			copy(ln.buf.data[ln.lenPS1:], anotherLine)

			if err = ln.buf.refresh(); err != nil {
				return "", err
			}
			continue
		case _LEFT:
			if _, err = ln.buf.backward(); err != nil {
				return "", err
			}
			continue
		case _RIGHT:
			if _, err = ln.buf.forward(); err != nil {
				return "", err
			}
			continue
		case _HOME:
			if err = ln.buf.start(); err != nil {
				return "", err
			}
			continue
		case _END:
			if _, err = ln.buf.end(); err != nil {
				return "", err
			}
			continue
		}
	}
}
