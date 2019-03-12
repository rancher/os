package tftp

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// NewServer creates TFTP server. It requires two functions to handle
// read and write requests.
// In case nil is provided for read or write handler the respective
// operation is disabled.
func NewServer(readHandler func(filename string, rf io.ReaderFrom) error,
	writeHandler func(filename string, wt io.WriterTo) error) *Server {
	return &Server{
		readHandler:  readHandler,
		writeHandler: writeHandler,
		timeout:      defaultTimeout,
		retries:      defaultRetries,
	}
}

type Server struct {
	readHandler  func(filename string, rf io.ReaderFrom) error
	writeHandler func(filename string, wt io.WriterTo) error
	backoff      backoffFunc
	conn         *net.UDPConn
	quit         chan chan struct{}
	wg           sync.WaitGroup
	timeout      time.Duration
	retries      int
}

// SetTimeout sets maximum time server waits for single network
// round-trip to succeed.
// Default is 5 seconds.
func (s *Server) SetTimeout(t time.Duration) {
	if t <= 0 {
		s.timeout = defaultTimeout
	} else {
		s.timeout = t
	}
}

// SetRetries sets maximum number of attempts server made to transmit a
// packet.
// Default is 5 attempts.
func (s *Server) SetRetries(count int) {
	if count < 1 {
		s.retries = defaultRetries
	} else {
		s.retries = count
	}
}

// SetBackoff sets a user provided function that is called to provide a
// backoff duration prior to retransmitting an unacknowledged packet.
func (s *Server) SetBackoff(h backoffFunc) {
	s.backoff = h
}

// ListenAndServe binds to address provided and start the server.
// ListenAndServe returns when Shutdown is called.
func (s *Server) ListenAndServe(addr string) error {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", a)
	if err != nil {
		return err
	}
	s.Serve(conn)
	return nil
}

// Serve starts server provided already opened UDP connecton. It is
// useful for the case when you want to run server in separate goroutine
// but still want to be able to handle any errors opening connection.
// Serve returns when Shutdown is called or connection is closed.
func (s *Server) Serve(conn *net.UDPConn) {
	s.conn = conn
	s.quit = make(chan chan struct{})
	for {
		select {
		case q := <-s.quit:
			q <- struct{}{}
			return
		default:
			err := s.processRequest(s.conn)
			if err != nil {
				// TODO: add logging handler
			}
		}
	}
}

// Shutdown make server stop listening for new requests, allows
// server to finish outstanding transfers and stops server.
func (s *Server) Shutdown() {
	s.conn.Close()
	q := make(chan struct{})
	s.quit <- q
	<-q
	s.wg.Wait()
}

func (s *Server) processRequest(conn *net.UDPConn) error {
	var buffer []byte
	buffer = make([]byte, datagramLength)
	n, remoteAddr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return fmt.Errorf("reading UDP: %v", err)
	}
	p, err := parsePacket(buffer[:n])
	if err != nil {
		return err
	}
	switch p := p.(type) {
	case pWRQ:
		filename, mode, opts, err := unpackRQ(p)
		if err != nil {
			return fmt.Errorf("unpack WRQ: %v", err)
		}
		//fmt.Printf("got WRQ (filename=%s, mode=%s, opts=%v)\n", filename, mode, opts)
		conn, err := net.ListenUDP("udp", &net.UDPAddr{})
		if err != nil {
			return err
		}
		if err != nil {
			return fmt.Errorf("open transmission: %v", err)
		}
		wt := &receiver{
			send:    make([]byte, datagramLength),
			receive: make([]byte, datagramLength),
			conn:    conn,
			retry:   &backoff{handler: s.backoff},
			timeout: s.timeout,
			retries: s.retries,
			addr:    remoteAddr,
			mode:    mode,
			opts:    opts,
		}
		s.wg.Add(1)
		go func() {
			if s.writeHandler != nil {
				err := s.writeHandler(filename, wt)
				if err != nil {
					wt.abort(err)
				} else {
					wt.terminate()
					wt.conn.Close()
				}
			} else {
				wt.abort(fmt.Errorf("server does not support write requests"))
			}
			s.wg.Done()
		}()
	case pRRQ:
		filename, mode, opts, err := unpackRQ(p)
		if err != nil {
			return fmt.Errorf("unpack RRQ: %v", err)
		}
		//fmt.Printf("got RRQ (filename=%s, mode=%s, opts=%v)\n", filename, mode, opts)
		conn, err := net.ListenUDP("udp", &net.UDPAddr{})
		if err != nil {
			return err
		}
		rf := &sender{
			send:    make([]byte, datagramLength),
			receive: make([]byte, datagramLength),
			tid:     remoteAddr.Port,
			conn:    conn,
			retry:   &backoff{handler: s.backoff},
			timeout: s.timeout,
			retries: s.retries,
			addr:    remoteAddr,
			mode:    mode,
			opts:    opts,
		}
		s.wg.Add(1)
		go func() {
			if s.readHandler != nil {
				err := s.readHandler(filename, rf)
				if err != nil {
					rf.abort(err)
				}
			} else {
				rf.abort(fmt.Errorf("server does not support read requests"))
			}
			s.wg.Done()
		}()
	default:
		return fmt.Errorf("unexpected %T", p)
	}
	return nil
}
