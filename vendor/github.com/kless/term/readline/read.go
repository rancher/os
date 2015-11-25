// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package readline

import "github.com/kless/term"

// Default values for prompts.
const (
	PS1 = "$ "
	PS2 = "> "
)

// keyAction represents the action to run for a key or sequence of keys pressed.
type keyAction int

const (
	_ keyAction = iota
	_LEFT
	_RIGHT
	_UP
	_DOWN

	_HOME
	_END
)

// To detect if has been pressed CTRL-C
var ChanCtrlC = make(chan byte)

// To detect if has been pressed CTRL-D
var ChanCtrlD = make(chan byte)

// A Line represents a line in the term.
type Line struct {
	ter  *term.Terminal
	buf  *buffer  // Text buffer
	hist *history // History file

	ps1    string // Primary prompt
	ps2    string // Command continuations
	lenPS1 int    // Size of primary prompt

	useHistory bool
}

// NewDefaultLine returns a line type using the prompt by default, and setting
// the terminal to raw mode.
// If the history is nil then it is not used.
func NewDefaultLine(hist *history) (*Line, error) {
	ter, err := term.New()
	if err != nil {
		return nil, err
	}
	if err = ter.RawMode(); err != nil {
		return nil, err
	}

	_, col, err := ter.GetSize()
	if err != nil {
		return nil, err
	}

	buf := newBuffer(len(PS1), col)
	buf.insertRunes([]rune(PS1))

	return &Line{
		ter: ter,
		buf: buf,
		hist: hist,

		ps1: PS1,
		ps2: PS2,
		lenPS1: len(PS1),

		useHistory: hasHistory(hist),
	}, nil
}

// Restore restores the terminal settings, so it is disabled the raw mode.
func (ln *Line) Restore() error {
	return ln.ter.Restore()
}
