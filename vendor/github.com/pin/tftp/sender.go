package tftp

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/pin/tftp/netascii"
)

// OutgoingTransfer provides methods to set the outgoing transfer size and
// retrieve the remote address of the peer.
type OutgoingTransfer interface {
	// SetSize is used to set the outgoing transfer size (tsize option: RFC2349)
	// manually in a server write transfer handler.
	//
	// It is not necessary in most cases; when the io.Reader provided to
	// ReadFrom also satisfies io.Seeker (e.g. os.File) the transfer size will
	// be determined automatically. Seek will not be attempted when the
	// transfer size option is set with SetSize.
	//
	// The value provided will be used only if SetSize is called before ReadFrom
	// and only on in a server read handler.
	SetSize(n int64)

	// RemoteAddr returns the remote peer's IP address and port.
	RemoteAddr() net.UDPAddr
}

type sender struct {
	conn    *net.UDPConn
	addr    *net.UDPAddr
	tid     int
	send    []byte
	receive []byte
	retry   *backoff
	timeout time.Duration
	retries int
	block   uint16
	mode    string
	opts    options
}

func (s *sender) RemoteAddr() net.UDPAddr { return *s.addr }

func (s *sender) SetSize(n int64) {
	if s.opts != nil {
		if _, ok := s.opts["tsize"]; ok {
			s.opts["tsize"] = strconv.FormatInt(n, 10)
		}
	}
}

func (s *sender) ReadFrom(r io.Reader) (n int64, err error) {
	if s.mode == "netascii" {
		r = netascii.ToReader(r)
	}
	if s.opts != nil {
		// check that tsize is set
		if ts, ok := s.opts["tsize"]; ok {
			// check that tsize is not set with SetSize already
			i, err := strconv.ParseInt(ts, 10, 64)
			if err == nil && i == 0 {
				if rs, ok := r.(io.Seeker); ok {
					pos, err := rs.Seek(0, 1)
					if err != nil {
						return 0, err
					}
					size, err := rs.Seek(0, 2)
					if err != nil {
						return 0, err
					}
					s.opts["tsize"] = strconv.FormatInt(size, 10)
					_, err = rs.Seek(pos, 0)
					if err != nil {
						return 0, err
					}
				}
			}
		}
		err = s.sendOptions()
		if err != nil {
			s.abort(err)
			return 0, err
		}
	}
	s.block = 1 // start data transmission with block 1
	binary.BigEndian.PutUint16(s.send[0:2], opDATA)
	for {
		l, err := io.ReadFull(r, s.send[4:])
		n += int64(l)
		if err != nil && err != io.ErrUnexpectedEOF {
			if err == io.EOF {
				binary.BigEndian.PutUint16(s.send[2:4], s.block)
				_, err = s.sendWithRetry(4)
				if err != nil {
					s.abort(err)
					return n, err
				}
				s.conn.Close()
				return n, nil
			}
			s.abort(err)
			return n, err
		}
		binary.BigEndian.PutUint16(s.send[2:4], s.block)
		_, err = s.sendWithRetry(4 + l)
		if err != nil {
			s.abort(err)
			return n, err
		}
		if l < len(s.send)-4 {
			s.conn.Close()
			return n, nil
		}
		s.block++
	}
}

func (s *sender) sendOptions() error {
	for name, value := range s.opts {
		if name == "blksize" {
			err := s.setBlockSize(value)
			if err != nil {
				delete(s.opts, name)
				continue
			}
		} else if name == "tsize" {
			if value != "0" {
				s.opts["tsize"] = value
			} else {
				delete(s.opts, name)
				continue
			}
		} else {
			delete(s.opts, name)
		}
	}
	if len(s.opts) > 0 {
		m := packOACK(s.send, s.opts)
		_, err := s.sendWithRetry(m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *sender) setBlockSize(blksize string) error {
	n, err := strconv.Atoi(blksize)
	if err != nil {
		return err
	}
	if n < 512 {
		return fmt.Errorf("blkzise too small: %d", n)
	}
	if n > 65464 {
		return fmt.Errorf("blksize too large: %d", n)
	}
	s.send = make([]byte, n+4)
	return nil
}

func (s *sender) sendWithRetry(l int) (*net.UDPAddr, error) {
	s.retry.reset()
	for {
		addr, err := s.sendDatagram(l)
		if _, ok := err.(net.Error); ok && s.retry.count() < s.retries {
			s.retry.backoff()
			continue
		}
		return addr, err
	}
}

func (s *sender) sendDatagram(l int) (*net.UDPAddr, error) {
	err := s.conn.SetReadDeadline(time.Now().Add(s.timeout))
	if err != nil {
		return nil, err
	}
	_, err = s.conn.WriteToUDP(s.send[:l], s.addr)
	if err != nil {
		return nil, err
	}
	for {
		n, addr, err := s.conn.ReadFromUDP(s.receive)
		if err != nil {
			return nil, err
		}
		if !addr.IP.Equal(s.addr.IP) || (s.tid != 0 && addr.Port != s.tid) {
			continue
		}
		p, err := parsePacket(s.receive[:n])
		if err != nil {
			continue
		}
		s.tid = addr.Port
		switch p := p.(type) {
		case pACK:
			if p.block() == s.block {
				return addr, nil
			}
		case pOACK:
			opts, err := unpackOACK(p)
			if s.block != 0 {
				continue
			}
			if err != nil {
				s.abort(err)
				return addr, err
			}
			for name, value := range opts {
				if name == "blksize" {
					err := s.setBlockSize(value)
					if err != nil {
						continue
					}
				}
			}
			return addr, nil
		case pERROR:
			return nil, fmt.Errorf("sending block %d: code=%d, error: %s",
				s.block, p.code(), p.message())
		}
	}
}

func (s *sender) abort(err error) error {
	if s.conn == nil {
		return nil
	}
	n := packERROR(s.send, 1, err.Error())
	_, err = s.conn.WriteToUDP(s.send[:n], s.addr)
	if err != nil {
		return err
	}
	s.conn.Close()
	s.conn = nil
	return nil
}
