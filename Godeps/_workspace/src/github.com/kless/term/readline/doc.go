// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package readline provides simple functions for both line and screen editing.

Features:

   Unicode support
   History
   Multi-line editing

List of key sequences enabled (just like in GNU Readline):

   Backspace / Ctrl+h

   Delete
   Home / Ctrl+a
   End  / Ctrl+e

   Left arrow  / Ctrl+b
   Right arrow / Ctrl+f
   Up arrow    / Ctrl+p
   Down arrow  / Ctrl+n
   Ctrl+left arrow
   Ctrl+right arrow

   Ctrl+t : swap actual character by the previous one
   Ctrl+k : delete from current to end of line
   Ctrl+u : delete the whole line
   Ctrl+l : clear screen

   Ctrl+c
   Ctrl+d : exit

Note that There are several default values:

+ For the buffer: BufferCap, BufferLen.

+ For the history file: HistoryCap, HistoryPerm.

Important: the TTY is set in "raw mode" so there is to use CR+LF ("\r\n") for
writing a new line.

Note: the values for the input and output are got from the package base "term".
*/
package readline
