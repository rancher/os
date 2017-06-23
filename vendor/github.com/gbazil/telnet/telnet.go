// Package telnet provides very simple interface for interacting with telnet devices from go routines.
package telnet

import (
	"bufio"
	"bytes"
	"net"
	"strings"
	"time"
)

// Telnet presents struct with net.Conn interface for telnet protocol plus buffered reader and timeout setup
type Telnet struct {
	conn    net.Conn
	reader  *bufio.Reader
	timeout time.Duration
}

// Dial constructs connection to a telnet device. Address string must be in format: "ip:port" (e.g. "127.0.0.1:23").
// Default timeout is set to 5 seconds.
func Dial(addr string) (t Telnet, err error) {
	t.conn, err = net.Dial("tcp", addr)

	if err == nil {
		t.reader = bufio.NewReader(t.conn)
		t.timeout = time.Second * 5 // default
	}

	return
}

// DialTimeout acts like Dial but takes a specific timeout (in nanoseconds).
func DialTimeout(addr string, timeout time.Duration) (t Telnet, err error) {
	t.conn, err = net.DialTimeout("tcp", addr, timeout)

	if err == nil {
		t.reader = bufio.NewReader(t.conn)
		t.timeout = timeout
	}

	return
}

// Read reads all data into string from telnet device until it meets the expected or stops on timeout.
func (t Telnet) Read(expect string) (str string, err error) {
	var buf bytes.Buffer
	t.conn.SetReadDeadline(time.Now().Add(t.timeout))

	for {
		b, e := t.reader.ReadByte()
		if e != nil {
			err = e
			break
		}

		if b == 255 {
			t.reader.Discard(2)
		} else {
			buf.WriteByte(b)
		}

		if strings.Contains(buf.String(), expect) {
			str = buf.String()
			break
		}
	}

	return
}

// Write writes string (command or data) to telnet device. Do not forget add LF to end of string!
func (t Telnet) Write(s string) (i int, err error) {
	t.conn.SetWriteDeadline(time.Now().Add(t.timeout))
	i, err = t.conn.Write([]byte(s))
	return
}

// SetTimeout changes default or start timeout for all interactions
func (t Telnet) SetTimeout(timeout time.Duration) {
	t.timeout = timeout
}
