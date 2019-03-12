TFTP server and client library for Golang
=========================================

[![GoDoc](https://godoc.org/github.com/pin/tftp?status.svg)](https://godoc.org/github.com/pin/tftp)
[![Build Status](https://travis-ci.org/pin/tftp.svg?branch=master)](https://travis-ci.org/pin/tftp)

Implements:
 * [RFC 1350](https://tools.ietf.org/html/rfc1350) - The TFTP Protocol (Revision 2)
 * [RFC 2347](https://tools.ietf.org/html/rfc2347) - TFTP Option Extension
 * [RFC 2348](https://tools.ietf.org/html/rfc2348) - TFTP Blocksize Option

Partially implements (tsize server side only):
 * [RFC 2349](https://tools.ietf.org/html/rfc2349) - TFTP Timeout Interval and Transfer Size Options

Set of features is sufficient for PXE boot support.

``` go
import "github.com/pin/tftp"
```

The package is cohesive to Golang `io`. Particularly it implements
`io.ReaderFrom` and `io.WriterTo` interfaces. That allows efficient data
transmission without unnecessary memory copying and allocations.


TFTP Server
-----------

```go

// readHandler is called when client starts file download from server
func readHandler(filename string, rf io.ReaderFrom) error {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	n, err := rf.ReadFrom(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	fmt.Printf("%d bytes sent\n", n)
	return nil
}

// writeHandler is called when client starts file upload to server
func writeHandler(filename string, wt io.WriterTo) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	n, err := wt.WriteTo(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	fmt.Printf("%d bytes received\n", n)
	return nil
}

func main() {
	// use nil in place of handler to disable read or write operations
	s := tftp.NewServer(readHandler, writeHandler)
	s.SetTimeout(5 * time.Second) // optional
	err := s.ListenAndServe(":69") // blocks until s.Shutdown() is called
	if err != nil {
		fmt.Fprintf(os.Stdout, "server: %v\n", err)
		os.Exit(1)
	}
}
```

TFTP Client
-----------
Upload file to server:

```go
c, err := tftp.NewClient("172.16.4.21:69")
file, err := os.Open(path)
c.SetTimeout(5 * time.Second) // optional
rf, err := c.Send("foobar.txt", "octet")
n, err := rf.ReadFrom(file)
fmt.Printf("%d bytes sent\n", n)
```

Download file from server:

```go
c, err := tftp.NewClient("172.16.4.21:69")
wt, err := c.Receive("foobar.txt", "octet")
file, err := os.Create(path)
// Optionally obtain transfer size before actual data.
if n, ok := wt.(IncomingTransfer).Size(); ok {
	fmt.Printf("Transfer size: %d\n", n)
}
n, err := wt.WriteTo(file)
fmt.Printf("%d bytes received\n", n)
```

Note: please handle errors better :)

TSize option
------------

PXE boot ROM often expects tsize option support from a server: client
(e.g. computer that boots over the network) wants to know size of a
download before the actual data comes. Server has to obtain stream
size and send it to a client.

Often it will happen automatically because TFTP library tries to check
if `io.Reader` provided to `ReadFrom` method also satisfies
`io.Seeker` interface (`os.File` for instance) and uses `Seek` to
determine file size.

In case `io.Reader` you provide to `ReadFrom` in read handler does not
satisfy `io.Seeker` interface or you do not want TFTP library to call
`Seek` on your reader but still want to respond with tsize option
during outgoing request you can use an `OutgoingTransfer` interface:

```go

func readHandler(filename string, rf io.ReaderFrom) error {
	...
	// Set transfer size before calling ReadFrom.
	rf.(tftp.OutgoingTransfer).SetSize(myFileSize)
	...
	// ReadFrom ...

```

Similarly, it is possible to obtain size of a file that is about to be
received using `IncomingTransfer` interface (see `Size` method).

Remote Address
--------------

The `OutgoingTransfer` and `IncomingTransfer` interfaces also provide the
`RemoteAddr` method which returns the peer IP address and port as a
`net.UDPAddr`.  This can be used for detailed logging in a server handler.

```go

func readHandler(filename string, rf io.ReaderFrom) error {
        ...
        raddr := rf.(tftp.OutgoingTransfer).RemoteAddr()
        log.Println("RRQ from", raddr.String())
        ...
        // ReadFrom ...
```

Backoff
-------

The default backoff before retransmitting an unacknowledged packet is a
random duration between 0 and 1 second.  This behavior can be overridden
in clients and servers by providing a custom backoff calculation function.

```go
	s := tftp.NewServer(readHandler, writeHandler)
	s.SetBackoff(func (attempts int) time.Duration {
		return time.Duration(attempts) * time.Second
	})
```

or, for no backoff

```go
	s.SetBackoff(func (int) time.Duration { return 0 })
```
