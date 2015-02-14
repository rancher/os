// Copyright 2012 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package term

// If it is true, at reading a password it shows characters shadowed by each key
// pressed.
var PasswordShadowed bool

var (
	_CTRL_C      = []byte{'^', 'C', '\r', '\n'}
	_RETURN      = []byte{'\r', '\n'}
	_SHADOW_CHAR = []byte{'*'}
)

// * * *

type modeType int

const (
	_ modeType = 1 << iota
	RawMode
	EchoMode
	CharMode
	PasswordMode
	OtherMode
)

// Mode returns the mode set in the terminal, if any.
func (t *Terminal) Mode() modeType {
	return t.mode
}
