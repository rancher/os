package util

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	bs       = []byte("\b \b")
	mask     = []byte("*")
	maxBytes = 512
)

func PromptPassword() (string, bool, error) {
	fmt.Print("Please enter password for [root]: ")
	p, err := MaskPassword(os.Stdin, os.Stdout)
	if err != nil {
		return "", false, errors.Wrapf(err, "failed to set password")
	}
	if string(p) == "" {
		fmt.Printf("Not setting password, leaving root disabled\n")
		return "", true, nil
	}
	fmt.Print("Confirm password for [root]: ")
	c, err := MaskPassword(os.Stdin, os.Stdout)
	if err != nil {
		return "", false, errors.Wrapf(err, "failed to confirm password")
	}
	return string(c), bytes.Compare(p, c) == 0, nil
}

func MaskPassword(r *os.File, w io.Writer) ([]byte, error) {
	var p []byte
	var err error
	fd := int(r.Fd())
	if terminal.IsTerminal(fd) {
		s, e := terminal.MakeRaw(fd)
		if e != nil {
			return p, e
		}
		defer func() {
			terminal.Restore(fd, s)
			fmt.Fprintln(w)
		}()
	}
	// Reference: ascii-table-0-127
	var i int
	for i = 0; i <= maxBytes; i++ {
		if v, e := getCharacter(r); e != nil {
			err = e
			break
		} else if v == 127 || v == 8 {
			// Delete || Backspace
			if l := len(p); l > 0 {
				p = p[:l-1]
				fmt.Fprint(w, string(bs))
			}
		} else if v == 13 || v == 10 {
			// CR || LF
			break
		} else if v == 3 {
			// End
			err = fmt.Errorf("interrupted")
			break
		} else if v != 0 {
			p = append(p, v)
			fmt.Fprint(w, string(mask))
		}
	}
	if i > maxBytes {
		err = fmt.Errorf("maximum password length is %v bytes", maxBytes)
	}
	return p, err
}

func getCharacter(r io.Reader) (byte, error) {
	buf := make([]byte, 1)
	if n, err := r.Read(buf); n == 0 || err != nil {
		if err != nil {
			return 0, err
		}
		return 0, io.EOF
	}
	return buf[0], nil
}
