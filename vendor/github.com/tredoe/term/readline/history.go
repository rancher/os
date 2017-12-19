// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package readline

import (
	"bufio"
	"container/list"
	"errors"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

// Values by default
var (
	HistoryCap              = 500  // Capacity
	HistoryPerm os.FileMode = 0600 // History file permission
)

var (
	ErrEmptyHist  = errors.New("history: empty")
	ErrNilElement = errors.New("history: no more elements")
)

// == Type

type history struct {
	Cap      int
	filename string
	file     *os.File
	mark     *list.Element // Pointer to the last element added.
	li       *list.List
}

// _baseHistory is the base to create an history file.
func _baseHistory(fname string, size int) (*history, error) {
	file, err := os.OpenFile(fname, os.O_CREATE|os.O_RDWR, HistoryPerm)
	if err != nil {
		return nil, err
	}

	h := new(history)
	h.Cap = size
	h.filename = fname
	h.file = file
	h.li = list.New()

	return h, nil
}

// NewHistory creates a new history using the maximum length by default.
func NewHistory(filename string) (*history, error) {
	return _baseHistory(filename, HistoryCap)
}

// NewHistoryOfSize creates a new history whose buffer has the specified size,
// which must be greater than zero.
func NewHistoryOfSize(filename string, size int) (*history, error) {
	if size <= 0 {
		return nil, errors.New("wrong history size: " + strconv.Itoa(size))
	}

	return _baseHistory(filename, size)
}

// == Access to file

// Load loads the history from the file.
func (h *history) Load() {
	in := bufio.NewReader(h.file)

	for {
		line, err := in.ReadString('\n')
		if err == io.EOF {
			break
		}

		h.li.PushBack(strings.TrimRight(line, "\n"))
	}

	h.mark = h.li.Back() // Point to an element.
}

// Save saves all lines to the text file, excep when:
// + it starts with some space
// + it is an empty line
func (h *history) Save() (err error) {
	if _, err = h.file.Seek(0, 0); err != nil {
		return
	}

	out := bufio.NewWriter(h.file)
	element := h.li.Front() // Get the first element.

	for i := 0; i < h.li.Len(); i++ {
		line := element.Value.(string)

		if strings.HasPrefix(line, " ") {
			goto _next
		}
		if line = strings.TrimSpace(line); line == "" {
			goto _next
		}
		if _, err = out.WriteString(line + "\n"); err != nil {
			log.Println("history.Save:", err)
			break
		}

	_next:
		if element = element.Next(); element == nil {
			continue
		}
	}

	if err = out.Flush(); err != nil {
		log.Println("history.Save:", err)
	}

	h.close()
	return
}

// close Close the file descriptor.
func (h *history) close() {
	h.file.Close()
}

// openFile opens again the file.
/*func (h *history) openFile() {
	file, err := os.Open(fname, os.O_CREATE|os.O_RDWR, HistoryPerm)
	if err != nil {
		log.Println("history.openFile:", err)
		return
	}

	h.file = file
}*/

// Add adds a new line to the buffer.
func (h *history) Add(line string) {
	if h.li.Len() <= h.Cap {
		h.mark = h.li.PushBack(line)
	} else {
		// TODO: overwrite lines
	}
}

// _baseNextPrev is the base to move between lines.
func (h *history) _baseNextPrev(c byte) (line []rune, err error) {
	if h.li.Len() <= 0 {
		return line, ErrEmptyHist
	}

	new := new(list.Element)
	if c == 'p' {
		new = h.mark.Prev()
	} else if c == 'n' {
		new = h.mark.Next()
	} else {
		panic("history._baseNextPrev: wrong character choice")
	}

	if new != nil {
		h.mark = new
	} else {
		return nil, ErrNilElement
	}

	return []rune(new.Value.(string)), nil
}

// Next returns the next line.
func (h *history) Next() (line []rune, err error) {
	return h._baseNextPrev('n')
}

// Prev returns the previous line.
func (h *history) Prev() (line []rune, err error) {
	return h._baseNextPrev('p')
}

// == Utility

// hasHistory checks whether has an history file.
func hasHistory(h *history) bool {
	if h == nil {
		return false
	}
	return true
}
