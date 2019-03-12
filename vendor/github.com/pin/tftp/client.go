package tftp

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

// NewClient creates TFTP client for server on address provided.
func NewClient(addr string) (*Client, error) {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("resolving address %s: %v", addr, err)
	}
	return &Client{
		addr:    a,
		timeout: defaultTimeout,
		retries: defaultRetries,
	}, nil
}

// SetTimeout sets maximum time client waits for single network round-trip to succeed.
// Default is 5 seconds.
func (c *Client) SetTimeout(t time.Duration) {
	if t <= 0 {
		c.timeout = defaultTimeout
	}
	c.timeout = t
}

// SetRetries sets maximum number of attempts client made to transmit a packet.
// Default is 5 attempts.
func (c *Client) SetRetries(count int) {
	if count < 1 {
		c.retries = defaultRetries
	}
	c.retries = count
}

// SetBackoff sets a user provided function that is called to provide a
// backoff duration prior to retransmitting an unacknowledged packet.
func (c *Client) SetBackoff(h backoffFunc) {
	c.backoff = h
}

type Client struct {
	addr    *net.UDPAddr
	timeout time.Duration
	retries int
	backoff backoffFunc
	blksize int
	tsize   bool
}

// Send starts outgoing file transmission. It returns io.ReaderFrom or error.
func (c Client) Send(filename string, mode string) (io.ReaderFrom, error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{})
	if err != nil {
		return nil, err
	}
	s := &sender{
		send:    make([]byte, datagramLength),
		receive: make([]byte, datagramLength),
		conn:    conn,
		retry:   &backoff{handler: c.backoff},
		timeout: c.timeout,
		retries: c.retries,
		addr:    c.addr,
		mode:    mode,
	}
	if c.blksize != 0 {
		s.opts = make(options)
		s.opts["blksize"] = strconv.Itoa(c.blksize)
	}
	n := packRQ(s.send, opWRQ, filename, mode, s.opts)
	addr, err := s.sendWithRetry(n)
	if err != nil {
		return nil, err
	}
	s.addr = addr
	s.opts = nil
	return s, nil
}

// Receive starts incoming file transmission. It returns io.WriterTo or error.
func (c Client) Receive(filename string, mode string) (io.WriterTo, error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{})
	if err != nil {
		return nil, err
	}
	if c.timeout == 0 {
		c.timeout = defaultTimeout
	}
	r := &receiver{
		send:     make([]byte, datagramLength),
		receive:  make([]byte, datagramLength),
		conn:     conn,
		retry:    &backoff{handler: c.backoff},
		timeout:  c.timeout,
		retries:  c.retries,
		addr:     c.addr,
		autoTerm: true,
		block:    1,
		mode:     mode,
	}
	if c.blksize != 0 || c.tsize {
		r.opts = make(options)
	}
	if c.blksize != 0 {
		r.opts["blksize"] = strconv.Itoa(c.blksize)
	}
	if c.tsize {
		r.opts["tsize"] = "0"
	}
	n := packRQ(r.send, opRRQ, filename, mode, r.opts)
	l, addr, err := r.receiveWithRetry(n)
	if err != nil {
		return nil, err
	}
	r.l = l
	r.addr = addr
	return r, nil
}
