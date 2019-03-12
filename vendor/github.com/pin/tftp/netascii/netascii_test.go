package netascii

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
	"testing/iotest"
)

var basic = map[string]string{
	"\r":     "\r\x00",
	"\n":     "\r\n",
	"la\nbu": "la\r\nbu",
	"la\rbu": "la\r\x00bu",
	"\r\r\r": "\r\x00\r\x00\r\x00",
	"\n\n\n": "\r\n\r\n\r\n",
}

func TestTo(t *testing.T) {
	for text, netascii := range basic {
		to := ToReader(strings.NewReader(text))
		n, _ := ioutil.ReadAll(to)
		if bytes.Compare(n, []byte(netascii)) != 0 {
			t.Errorf("%q to netascii: %q != %q", text, n, netascii)
		}
	}
}

func TestFrom(t *testing.T) {
	for text, netascii := range basic {
		r := bytes.NewReader([]byte(netascii))
		b := &bytes.Buffer{}
		from := FromWriter(b)
		r.WriteTo(from)
		n, _ := ioutil.ReadAll(b)
		if string(n) != text {
			t.Errorf("%q from netascii: %q != %q", netascii, n, text)
		}
	}
}

const text = `
Therefore, the sequence "CR LF" must be treated as a single "new
line" character and used whenever their combined action is
intended; the sequence "CR NUL" must be used where a carriage
return alone is actually desired; and the CR character must be
avoided in other contexts.  This rule gives assurance to systems
which must decide whether to perform a "new line" function or a
multiple-backspace that the TELNET stream contains a character
following a CR that will allow a rational decision.
(in the default ASCII mode), to preserve the symmetry of the
NVT model.  Even though it may be known in some situations
(e.g., with remote echo and suppress go ahead options in
effect) that characters are not being sent to an actual
printer, nonetheless, for the sake of consistency, the protocol
requires that a NUL be inserted following a CR not followed by
a LF in the data stream.  The converse of this is that a NUL
received in the data stream after a CR (in the absence of
options negotiations which explicitly specify otherwise) should
be stripped out prior to applying the NVT to local character
set mapping.
`

func TestWriteRead(t *testing.T) {
	var one bytes.Buffer
	to := ToReader(strings.NewReader(text))
	one.ReadFrom(to)
	two := &bytes.Buffer{}
	from := FromWriter(two)
	one.WriteTo(from)
	text2, _ := ioutil.ReadAll(two)
	if text != string(text2) {
		t.Errorf("text mismatch \n%x \n%x", text, text2)
	}
}

func TestOneByte(t *testing.T) {
	var one bytes.Buffer
	to := iotest.OneByteReader(ToReader(strings.NewReader(text)))
	one.ReadFrom(to)
	two := &bytes.Buffer{}
	from := FromWriter(two)
	one.WriteTo(from)
	text2, _ := ioutil.ReadAll(two)
	if text != string(text2) {
		t.Errorf("text mismatch \n%x \n%x", text, text2)
	}
}
