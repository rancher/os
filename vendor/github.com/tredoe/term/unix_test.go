// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// +build !plan9,!windows

package term

import (
	"syscall"
	"testing"
)

func init() {
	InputFD = syscall.Stderr
}

func TestRawMode(t *testing.T) {
	ter, err := New()
	if err != nil {
		t.Fatal(err)
	}

	oldState := ter.oldState

	if err = ter.RawMode(); err != nil {
		t.Error("expected set raw mode:", err)
	}
	if err = ter.Restore(); err != nil {
		t.Error("expected to restore:", err)
	}

	lastState := ter.lastState

	if oldState.Iflag != lastState.Iflag ||
		oldState.Oflag != lastState.Oflag ||
		oldState.Cflag != lastState.Cflag ||
		oldState.Lflag != lastState.Lflag {

		t.Error("expected to restore all settings")
	}

	// Restore from a saved state
	ter, _ = New()
	state := ter.OriginalState()

	if err = Restore(InputFD, state); err != nil {
		t.Error("expected to restore from saved state:", err)
	}
}

func TestInformation(t *testing.T) {
	if !SupportANSI() {
		t.Error("expected to support this terminal")
	}
	if !IsTerminal(InputFD) {
		t.Error("expected to be a terminal")
	}

	/*ter, _ := New()
	if _, err := TTYName(ter.fd); err != nil {
		t.Error("expected to get the terminal name", err)
	}
	ter.Restore()*/
}

func TestSize(t *testing.T) {
	ter, _ := New()
	defer ter.Restore()

	row, col, err := ter.GetSize()
	if err != nil {
		t.Error(err)
		return
	}
	if row == 0 || col == 0 {
		t.Error("expected to get size")
	}

	//rowE, colE := GetSizeFromEnv()
	//if rowE == 0 || colE == 0 {
		//t.Error("expected to get size from environment")
	//}
}
