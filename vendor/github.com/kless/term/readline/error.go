// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package readline

import "errors"

var ErrCtrlD = errors.New("Interrumpted (Ctrl+d)")

// An inputError represents a failure on input.
type inputError string

func (e inputError) Error() string {
	return "could not read from input: " + string(e)
}

// An outputError represents a failure in output.
type outputError string

func (e outputError) Error() string {
	return "could not write to output: " + string(e)
}
