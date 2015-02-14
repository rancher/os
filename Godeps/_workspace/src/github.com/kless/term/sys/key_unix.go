// Copyright 2013 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// +build !plan9,!windows

package sys

// Key codes
const (
	K_TAB    = 0x09 // TAB key
	K_BACK   = 0x7F // BACKSPACE key
	K_RETURN = 0x0D // RETURN key
	K_ESCAPE = 0x1B // ESC key
)

// Control+letters key codes.
const (
	K_CTRL_A = iota + 0x01
	K_CTRL_B
	K_CTRL_C
	K_CTRL_D
	K_CTRL_E
	K_CTRL_F
	K_CTRL_G
	K_CTRL_H
	K_CTRL_I
	K_CTRL_J
	K_CTRL_K
	K_CTRL_L
	K_CTRL_M
	K_CTRL_N
	K_CTRL_O
	K_CTRL_P
	K_CTRL_Q
	K_CTRL_R
	K_CTRL_S
	K_CTRL_T
	K_CTRL_U
	K_CTRL_V
	K_CTRL_W
	K_CTRL_X
	K_CTRL_Y
	K_CTRL_Z
)
