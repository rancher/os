package netascii

// TODO: make it work not only on linux

import "io"

const (
	CR  = '\x0d'
	LF  = '\x0a'
	NUL = '\x00'
)

func ToReader(r io.Reader) io.Reader {
	return &toReader{
		r:   r,
		buf: make([]byte, 256),
	}
}

type toReader struct {
	r   io.Reader
	buf []byte
	n   int
	i   int
	err error
	lf  bool
	nul bool
}

func (r *toReader) Read(p []byte) (int, error) {
	var n int
	for n < len(p) {
		if r.lf {
			p[n] = LF
			n++
			r.lf = false
			continue
		}
		if r.nul {
			p[n] = NUL
			n++
			r.nul = false
			continue
		}
		if r.i < r.n {
			if r.buf[r.i] == LF {
				p[n] = CR
				r.lf = true
			} else if r.buf[r.i] == CR {
				p[n] = CR
				r.nul = true

			} else {
				p[n] = r.buf[r.i]
			}
			r.i++
			n++
			continue
		}
		if r.err == nil {
			r.n, r.err = r.r.Read(r.buf)
			r.i = 0
		} else {
			return n, r.err
		}
	}
	return n, r.err
}

type fromWriter struct {
	w   io.Writer
	buf []byte
	i   int
	cr  bool
}

func FromWriter(w io.Writer) io.Writer {
	return &fromWriter{
		w:   w,
		buf: make([]byte, 256),
	}
}

func (w *fromWriter) Write(p []byte) (n int, err error) {
	for n < len(p) {
		if w.cr {
			if p[n] == LF {
				w.buf[w.i] = LF
			}
			if p[n] == NUL {
				w.buf[w.i] = CR
			}
			w.cr = false
			w.i++
		} else if p[n] == CR {
			w.cr = true
		} else {
			w.buf[w.i] = p[n]
			w.i++
		}
		n++
		if w.i == len(w.buf) || n == len(p) {
			_, err = w.w.Write(w.buf[:w.i])
			w.i = 0
		}
	}
	return n, err
}
