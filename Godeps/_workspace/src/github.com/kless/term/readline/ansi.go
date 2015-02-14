// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// The references about ANSI Escape sequences have been got from
// http://ascii-table.com/ansi-escape-sequences.php and
// http://www.termsys.demon.co.uk/vtansi.htm

package readline

// ANSI terminal escape controls
const (
	// Cursor control
	ANSI_CURSOR_UP       = "\033[A" // Up
	ANSI_CURSOR_DOWN     = "\033[B" // Down
	ANSI_CURSOR_FORWARD  = "\033[C" // Forward
	ANSI_CURSOR_BACKWARD = "\033[D" // Backward

	ANSI_NEXT_LINE = "\033[E" // To next line
	ANSI_PREV_LINE = "\033[F" // To previous line

	// Erase
	ANSI_DEL_LINE = "\033[2K" // Erase line

	// Graphics mode
	ANSI_SET_BOLD = "\033[1m" // Bold on
	ANSI_SET_OFF  = "\033[0m" // All attributes off
)

// ANSI terminal escape controls
var (
	// Cursor control
	CursorUp       = []byte(ANSI_CURSOR_UP)
	CursorDown     = []byte(ANSI_CURSOR_DOWN)
	CursorForward  = []byte(ANSI_CURSOR_FORWARD)
	CursorBackward = []byte(ANSI_CURSOR_BACKWARD)

	ToNextLine     = []byte(ANSI_NEXT_LINE)
	ToPreviousLine = []byte(ANSI_PREV_LINE)

	// Erase Text
	DelScreenToUpper = []byte("\033[2J\033[0;0H") // Erase the screen; move upper

	DelToRight       = []byte("\033[0K")       // Erase to right
	DelLine_CR       = []byte("\033[2K\r")     // Erase line; carriage return
	DelLine_cursorUp = []byte("\033[2K\033[A") // Erase line; cursor up

	//DelChar      = []byte("\033[1X") // Erase character
	DelChar      = []byte("\033[P") // Delete character, from current position
	DelBackspace = []byte("\033[D\033[P")

	// Misc.
	//InsertChar  = []byte("\033[@")   // Insert CHaracter
	//SetLineWrap = []byte("\033[?7h") // Enable Line Wrap
)

// Characters
var (
	CR    = []byte{13}     // Carriage return -- \r
	CRLF  = []byte{13, 10} // CR+LF is used for a new line in raw mode -- \r\n
	CtrlC = []rune("^C")
	CtrlD = []rune("^D")
)
