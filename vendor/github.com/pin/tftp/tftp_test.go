package tftp

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"testing"
	"testing/iotest"
	"time"
)

var localhost string = determineLocalhost()

func determineLocalhost() string {
	l, err := net.ListenTCP("tcp", nil)
	if err != nil {
		panic(fmt.Sprintf("ListenTCP error: %s", err))
	}
	_, lport, _ := net.SplitHostPort(l.Addr().String())
	defer l.Close()

	lo := make(chan string)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				break
			}
			conn.Close()
		}
	}()

	go func() {
		port, _ := strconv.Atoi(lport)
		for _, af := range []string{"tcp6", "tcp4"} {
			conn, err := net.DialTCP(af, &net.TCPAddr{}, &net.TCPAddr{Port: port})
			if err == nil {
				conn.Close()
				host, _, _ := net.SplitHostPort(conn.LocalAddr().String())
				lo <- host
				return
			}
		}
		panic("could not determine address family")
	}()

	return <-lo
}

func localSystem(c *net.UDPConn) string {
	_, port, _ := net.SplitHostPort(c.LocalAddr().String())
	return net.JoinHostPort(localhost, port)
}

func TestPackUnpack(t *testing.T) {
	v := []string{"test-filename/with-subdir"}
	testOptsList := []options{
		nil,
		options{
			"tsize":   "1234",
			"blksize": "22",
		},
	}
	for _, filename := range v {
		for _, mode := range []string{"octet", "netascii"} {
			for _, opts := range testOptsList {
				packUnpack(t, filename, mode, opts)
			}
		}
	}
}

func packUnpack(t *testing.T, filename, mode string, opts options) {
	b := make([]byte, datagramLength)
	for _, op := range []uint16{opRRQ, opWRQ} {
		n := packRQ(b, op, filename, mode, opts)
		f, m, o, err := unpackRQ(b[:n])
		if err != nil {
			t.Errorf("%s pack/unpack: %v", filename, err)
		}
		if f != filename {
			t.Errorf("filename mismatch (%s): '%x' vs '%x'",
				filename, f, filename)
		}
		if m != mode {
			t.Errorf("mode mismatch (%s): '%x' vs '%x'",
				mode, m, mode)
		}
		if opts != nil {
			for name, value := range opts {
				v, ok := o[name]
				if !ok {
					t.Errorf("missing %s option", name)
				}
				if v != value {
					t.Errorf("option %s mismatch: '%x' vs '%x'", name, v, value)
				}
			}
		}
	}
}

func TestZeroLength(t *testing.T) {
	s, c := makeTestServer()
	defer s.Shutdown()
	testSendReceive(t, c, 0)
}

func Test900(t *testing.T) {
	s, c := makeTestServer()
	defer s.Shutdown()
	for i := 600; i < 4000; i += 1 {
		c.blksize = i
		testSendReceive(t, c, 9000+int64(i))
	}
}

func Test1000(t *testing.T) {
	s, c := makeTestServer()
	defer s.Shutdown()
	for i := int64(0); i < 5000; i++ {
		filename := fmt.Sprintf("length-%d-bytes-%d", i, time.Now().UnixNano())
		rf, err := c.Send(filename, "octet")
		if err != nil {
			t.Fatalf("requesting %s write: %v", filename, err)
		}
		r := io.LimitReader(newRandReader(rand.NewSource(i)), i)
		n, err := rf.ReadFrom(r)
		if err != nil {
			t.Fatalf("sending %s: %v", filename, err)
		}
		if n != i {
			t.Errorf("%s length mismatch: %d != %d", filename, n, i)
		}
	}
}

func Test1810(t *testing.T) {
	s, c := makeTestServer()
	defer s.Shutdown()
	c.blksize = 1810
	testSendReceive(t, c, 9000+1810)
}

func TestTSize(t *testing.T) {
	s, c := makeTestServer()
	defer s.Shutdown()
	c.tsize = true
	testSendReceive(t, c, 640)
}

func TestNearBlockLength(t *testing.T) {
	s, c := makeTestServer()
	defer s.Shutdown()
	for i := 450; i < 520; i++ {
		testSendReceive(t, c, int64(i))
	}
}

func TestBlockWrapsAround(t *testing.T) {
	s, c := makeTestServer()
	defer s.Shutdown()
	n := 65535 * 512
	for i := n - 2; i < n+2; i++ {
		testSendReceive(t, c, int64(i))
	}
}

func TestRandomLength(t *testing.T) {
	s, c := makeTestServer()
	defer s.Shutdown()
	r := rand.New(rand.NewSource(42))
	for i := 0; i < 100; i++ {
		testSendReceive(t, c, r.Int63n(100000))
	}
}

func TestBigFile(t *testing.T) {
	s, c := makeTestServer()
	defer s.Shutdown()
	testSendReceive(t, c, 3*1000*1000)
}

func TestByOneByte(t *testing.T) {
	s, c := makeTestServer()
	defer s.Shutdown()
	filename := "test-by-one-byte"
	mode := "octet"
	const length = 80000
	sender, err := c.Send(filename, mode)
	if err != nil {
		t.Fatalf("requesting write: %v", err)
	}
	r := iotest.OneByteReader(io.LimitReader(
		newRandReader(rand.NewSource(42)), length))
	n, err := sender.ReadFrom(r)
	if err != nil {
		t.Fatalf("send error: %v", err)
	}
	if n != length {
		t.Errorf("%s read length mismatch: %d != %d", filename, n, length)
	}
	readTransfer, err := c.Receive(filename, mode)
	if err != nil {
		t.Fatalf("requesting read %s: %v", filename, err)
	}
	buf := &bytes.Buffer{}
	n, err = readTransfer.WriteTo(buf)
	if err != nil {
		t.Fatalf("%s read error: %v", filename, err)
	}
	if n != length {
		t.Errorf("%s read length mismatch: %d != %d", filename, n, length)
	}
	bs, _ := ioutil.ReadAll(io.LimitReader(
		newRandReader(rand.NewSource(42)), length))
	if !bytes.Equal(bs, buf.Bytes()) {
		t.Errorf("\nsent: %x\nrcvd: %x", bs, buf)
	}
}

func TestDuplicate(t *testing.T) {
	s, c := makeTestServer()
	defer s.Shutdown()
	filename := "test-duplicate"
	mode := "octet"
	bs := []byte("lalala")
	sender, err := c.Send(filename, mode)
	if err != nil {
		t.Fatalf("requesting write: %v", err)
	}
	buf := bytes.NewBuffer(bs)
	_, err = sender.ReadFrom(buf)
	if err != nil {
		t.Fatalf("send error: %v", err)
	}
	sender, err = c.Send(filename, mode)
	if err == nil {
		t.Fatalf("file already exists")
	}
	t.Logf("sending file that already exists: %v", err)
}

func TestNotFound(t *testing.T) {
	s, c := makeTestServer()
	defer s.Shutdown()
	filename := "test-not-exists"
	mode := "octet"
	_, err := c.Receive(filename, mode)
	if err == nil {
		t.Fatalf("file not exists", err)
	}
	t.Logf("receiving file that does not exist: %v", err)
}

func testSendReceive(t *testing.T, client *Client, length int64) {
	filename := fmt.Sprintf("length-%d-bytes", length)
	mode := "octet"
	writeTransfer, err := client.Send(filename, mode)
	if err != nil {
		t.Fatalf("requesting write %s: %v", filename, err)
	}
	r := io.LimitReader(newRandReader(rand.NewSource(42)), length)
	n, err := writeTransfer.ReadFrom(r)
	if err != nil {
		t.Fatalf("%s write error: %v", filename, err)
	}
	if n != length {
		t.Errorf("%s write length mismatch: %d != %d", filename, n, length)
	}
	readTransfer, err := client.Receive(filename, mode)
	if err != nil {
		t.Fatalf("requesting read %s: %v", filename, err)
	}
	if it, ok := readTransfer.(IncomingTransfer); ok {
		if n, ok := it.Size(); ok {
			fmt.Printf("Transfer size: %d\n", n)
			if n != length {
				t.Errorf("tsize mismatch: %d vs %d", n, length)
			}
		}
	}
	buf := &bytes.Buffer{}
	n, err = readTransfer.WriteTo(buf)
	if err != nil {
		t.Fatalf("%s read error: %v", filename, err)
	}
	if n != length {
		t.Errorf("%s read length mismatch: %d != %d", filename, n, length)
	}
	bs, _ := ioutil.ReadAll(io.LimitReader(
		newRandReader(rand.NewSource(42)), length))
	if !bytes.Equal(bs, buf.Bytes()) {
		t.Errorf("\nsent: %x\nrcvd: %x", bs, buf)
	}
}

func TestSendTsizeFromSeek(t *testing.T) {
	// create read-only server
	s := NewServer(func(filename string, rf io.ReaderFrom) error {
		b := make([]byte, 100)
		rr := newRandReader(rand.NewSource(42))
		rr.Read(b)
		// bytes.Reader implements io.Seek
		r := bytes.NewReader(b)
		_, err := rf.ReadFrom(r)
		if err != nil {
			t.Errorf("sending bytes: %v", err)
		}
		return nil
	}, nil)

	conn, err := net.ListenUDP("udp", &net.UDPAddr{})
	if err != nil {
		t.Fatalf("listening: %v", err)
	}

	go s.Serve(conn)
	defer s.Shutdown()

	c, _ := NewClient(localSystem(conn))
	c.tsize = true
	r, _ := c.Receive("f", "octet")
	var size int64
	if t, ok := r.(IncomingTransfer); ok {
		if n, ok := t.Size(); ok {
			size = n
			fmt.Printf("Transfer size: %d\n", n)
		}
	}

	if size != 100 {
		t.Errorf("size expected: 100, got %d", size)
	}

	r.WriteTo(ioutil.Discard)
}

type testBackend struct {
	m  map[string][]byte
	mu sync.Mutex
}

func makeTestServer() (*Server, *Client) {
	b := &testBackend{}
	b.m = make(map[string][]byte)

	// Create server
	s := NewServer(b.handleRead, b.handleWrite)

	conn, err := net.ListenUDP("udp", &net.UDPAddr{})
	if err != nil {
		panic(err)
	}

	go s.Serve(conn)

	// Create client for that server
	c, err := NewClient(localSystem(conn))
	if err != nil {
		panic(err)
	}

	return s, c
}

func TestNoHandlers(t *testing.T) {
	s := NewServer(nil, nil)

	conn, err := net.ListenUDP("udp", &net.UDPAddr{})
	if err != nil {
		panic(err)
	}

	go s.Serve(conn)

	c, err := NewClient(localSystem(conn))
	if err != nil {
		panic(err)
	}

	_, err = c.Send("test", "octet")
	if err == nil {
		t.Errorf("error expected")
	}

	_, err = c.Receive("test", "octet")
	if err == nil {
		t.Errorf("error expected")
	}
}

func (b *testBackend) handleWrite(filename string, wt io.WriterTo) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, ok := b.m[filename]
	if ok {
		fmt.Fprintf(os.Stderr, "File %s already exists\n", filename)
		return fmt.Errorf("file already exists")
	}
	if t, ok := wt.(IncomingTransfer); ok {
		if n, ok := t.Size(); ok {
			fmt.Printf("Transfer size: %d\n", n)
		}
	}
	buf := &bytes.Buffer{}
	_, err := wt.WriteTo(buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't receive %s: %v\n", filename, err)
		return err
	}
	b.m[filename] = buf.Bytes()
	return nil
}

func (b *testBackend) handleRead(filename string, rf io.ReaderFrom) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	bs, ok := b.m[filename]
	if !ok {
		fmt.Fprintf(os.Stderr, "File %s not found\n", filename)
		return fmt.Errorf("file not found")
	}
	if t, ok := rf.(OutgoingTransfer); ok {
		t.SetSize(int64(len(bs)))
	}
	_, err := rf.ReadFrom(bytes.NewBuffer(bs))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't send %s: %v\n", filename, err)
		return err
	}
	return nil
}

type randReader struct {
	src  rand.Source
	next int64
	i    int8
}

func newRandReader(src rand.Source) io.Reader {
	r := &randReader{
		src:  src,
		next: src.Int63(),
	}
	return r
}

func (r *randReader) Read(p []byte) (n int, err error) {
	next, i := r.next, r.i
	for n = 0; n < len(p); n++ {
		if i == 7 {
			next, i = r.src.Int63(), 0
		}
		p[n] = byte(next)
		next >>= 8
		i++
	}
	r.next, r.i = next, i
	return
}

func TestServerSendTimeout(t *testing.T) {
	s, c := makeTestServer()
	s.SetTimeout(time.Second)
	s.SetRetries(2)
	var serverErr error
	s.readHandler = func(filename string, rf io.ReaderFrom) error {
		r := io.LimitReader(newRandReader(rand.NewSource(42)), 80000)
		_, serverErr = rf.ReadFrom(r)
		return serverErr
	}
	defer s.Shutdown()
	filename := "test-server-send-timeout"
	mode := "octet"
	readTransfer, err := c.Receive(filename, mode)
	if err != nil {
		t.Fatalf("requesting read %s: %v", filename, err)
	}
	w := &slowWriter{
		n:     3,
		delay: 8 * time.Second,
	}
	_, _ = readTransfer.WriteTo(w)
	netErr, ok := serverErr.(net.Error)
	if !ok {
		t.Fatalf("network error expected: %T", serverErr)
	}
	if !netErr.Timeout() {
		t.Fatalf("timout is expected: %v", serverErr)
	}
}

func TestServerReceiveTimeout(t *testing.T) {
	s, c := makeTestServer()
	s.SetTimeout(time.Second)
	s.SetRetries(2)
	var serverErr error
	s.writeHandler = func(filename string, wt io.WriterTo) error {
		buf := &bytes.Buffer{}
		_, serverErr = wt.WriteTo(buf)
		return serverErr
	}
	defer s.Shutdown()
	filename := "test-server-receive-timeout"
	mode := "octet"
	writeTransfer, err := c.Send(filename, mode)
	if err != nil {
		t.Fatalf("requesting write %s: %v", filename, err)
	}
	r := &slowReader{
		r:     io.LimitReader(newRandReader(rand.NewSource(42)), 80000),
		n:     3,
		delay: 8 * time.Second,
	}
	_, _ = writeTransfer.ReadFrom(r)
	netErr, ok := serverErr.(net.Error)
	if !ok {
		t.Fatalf("network error expected: %T", serverErr)
	}
	if !netErr.Timeout() {
		t.Fatalf("timout is expected: %v", serverErr)
	}
}

func TestClientReceiveTimeout(t *testing.T) {
	s, c := makeTestServer()
	c.SetTimeout(time.Second)
	c.SetRetries(2)
	s.readHandler = func(filename string, rf io.ReaderFrom) error {
		r := &slowReader{
			r:     io.LimitReader(newRandReader(rand.NewSource(42)), 80000),
			n:     3,
			delay: 8 * time.Second,
		}
		_, err := rf.ReadFrom(r)
		return err
	}
	defer s.Shutdown()
	filename := "test-client-receive-timeout"
	mode := "octet"
	readTransfer, err := c.Receive(filename, mode)
	if err != nil {
		t.Fatalf("requesting read %s: %v", filename, err)
	}
	buf := &bytes.Buffer{}
	_, err = readTransfer.WriteTo(buf)
	netErr, ok := err.(net.Error)
	if !ok {
		t.Fatalf("network error expected: %T", err)
	}
	if !netErr.Timeout() {
		t.Fatalf("timout is expected: %v", err)
	}
}

func TestClientSendTimeout(t *testing.T) {
	s, c := makeTestServer()
	c.SetTimeout(time.Second)
	c.SetRetries(2)
	s.writeHandler = func(filename string, wt io.WriterTo) error {
		w := &slowWriter{
			n:     3,
			delay: 8 * time.Second,
		}
		_, err := wt.WriteTo(w)
		return err
	}
	defer s.Shutdown()
	filename := "test-client-send-timeout"
	mode := "octet"
	writeTransfer, err := c.Send(filename, mode)
	if err != nil {
		t.Fatalf("requesting write %s: %v", filename, err)
	}
	r := io.LimitReader(newRandReader(rand.NewSource(42)), 80000)
	_, err = writeTransfer.ReadFrom(r)
	netErr, ok := err.(net.Error)
	if !ok {
		t.Fatalf("network error expected: %T", err)
	}
	if !netErr.Timeout() {
		t.Fatalf("timout is expected: %v", err)
	}
}

type slowReader struct {
	r     io.Reader
	n     int64
	delay time.Duration
}

func (r *slowReader) Read(p []byte) (n int, err error) {
	if r.n > 0 {
		r.n--
		return r.r.Read(p)
	}
	time.Sleep(r.delay)
	return r.r.Read(p)
}

type slowWriter struct {
	r     io.Reader
	n     int64
	delay time.Duration
}

func (r *slowWriter) Write(p []byte) (n int, err error) {
	if r.n > 0 {
		r.n--
		return len(p), nil
	}
	time.Sleep(r.delay)
	return len(p), nil
}
