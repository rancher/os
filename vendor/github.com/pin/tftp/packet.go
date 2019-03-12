package tftp

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	opRRQ   = uint16(1) // Read request (RRQ)
	opWRQ   = uint16(2) // Write request (WRQ)
	opDATA  = uint16(3) // Data
	opACK   = uint16(4) // Acknowledgement
	opERROR = uint16(5) // Error
	opOACK  = uint16(6) // Options Acknowledgment
)

const (
	blockLength    = 512
	datagramLength = 516
)

type options map[string]string

// RRQ/WRQ packet
//
//  2 bytes     string    1 byte    string    1 byte
// --------------------------------------------------
// | Opcode |  Filename  |   0  |    Mode    |   0  |
// --------------------------------------------------
type pRRQ []byte
type pWRQ []byte

// packRQ returns length of the packet in b
func packRQ(p []byte, op uint16, filename, mode string, opts options) int {
	binary.BigEndian.PutUint16(p, op)
	n := 2
	n += copy(p[2:len(p)-10], filename)
	p[n] = 0
	n++
	n += copy(p[n:], mode)
	p[n] = 0
	n++
	for name, value := range opts {
		n += copy(p[n:], name)
		p[n] = 0
		n++
		n += copy(p[n:], value)
		p[n] = 0
		n++
	}
	return n
}

func unpackRQ(p []byte) (filename, mode string, opts options, err error) {
	bs := bytes.Split(p[2:], []byte{0})
	if len(bs) < 2 {
		return "", "", nil, fmt.Errorf("missing filename or mode")
	}
	filename = string(bs[0])
	mode = string(bs[1])
	if len(bs) < 4 {
		return filename, mode, nil, nil
	}
	opts = make(options)
	for i := 2; i+1 < len(bs); i += 2 {
		opts[string(bs[i])] = string(bs[i+1])
	}
	return filename, mode, opts, nil
}

// OACK packet
//
// +----------+---~~---+---+---~~---+---+---~~---+---+---~~---+---+
// |  Opcode  |  opt1  | 0 | value1 | 0 |  optN  | 0 | valueN | 0 |
// +----------+---~~---+---+---~~---+---+---~~---+---+---~~---+---+
type pOACK []byte

func packOACK(p []byte, opts options) int {
	binary.BigEndian.PutUint16(p, opOACK)
	n := 2
	for name, value := range opts {
		n += copy(p[n:], name)
		p[n] = 0
		n++
		n += copy(p[n:], value)
		p[n] = 0
		n++
	}
	return n
}

func unpackOACK(p []byte) (opts options, err error) {
	bs := bytes.Split(p[2:], []byte{0})
	opts = make(options)
	for i := 0; i+1 < len(bs); i += 2 {
		opts[string(bs[i])] = string(bs[i+1])
	}
	return opts, nil
}

// ERROR packet
//
//  2 bytes     2 bytes      string    1 byte
// ------------------------------------------
// | Opcode |  ErrorCode |   ErrMsg   |  0  |
// ------------------------------------------
type pERROR []byte

func packERROR(p []byte, code uint16, message string) int {
	binary.BigEndian.PutUint16(p, opERROR)
	binary.BigEndian.PutUint16(p[2:], code)
	n := copy(p[4:len(p)-2], message)
	p[4+n] = 0
	return n + 5
}

func (p pERROR) code() uint16 {
	return binary.BigEndian.Uint16(p[2:])
}

func (p pERROR) message() string {
	return string(p[4:])
}

// DATA packet
//
//  2 bytes    2 bytes     n bytes
// ----------------------------------
// | Opcode |   Block #  |   Data   |
// ----------------------------------
type pDATA []byte

func (p pDATA) block() uint16 {
	return binary.BigEndian.Uint16(p[2:])
}

// ACK packet
//
//  2 bytes    2 bytes
// -----------------------
// | Opcode |   Block #  |
// -----------------------
type pACK []byte

func (p pACK) block() uint16 {
	return binary.BigEndian.Uint16(p[2:])
}

func parsePacket(p []byte) (interface{}, error) {
	l := len(p)
	if l < 2 {
		return nil, fmt.Errorf("short packet")
	}
	opcode := binary.BigEndian.Uint16(p)
	switch opcode {
	case opRRQ:
		if l < 4 {
			return nil, fmt.Errorf("short RRQ packet: %d", l)
		}
		return pRRQ(p), nil
	case opWRQ:
		if l < 4 {
			return nil, fmt.Errorf("short WRQ packet: %d", l)
		}
		return pWRQ(p), nil
	case opDATA:
		if l < 4 {
			return nil, fmt.Errorf("short DATA packet: %d", l)
		}
		return pDATA(p), nil
	case opACK:
		if l < 4 {
			return nil, fmt.Errorf("short ACK packet: %d", l)
		}
		return pACK(p), nil
	case opERROR:
		if l < 5 {
			return nil, fmt.Errorf("short ERROR packet: %d", l)
		}
		return pERROR(p), nil
	case opOACK:
		if l < 6 {
			return nil, fmt.Errorf("short OACK packet: %d", l)
		}
		return pOACK(p), nil
	default:
		return nil, fmt.Errorf("unknown opcode: %d", opcode)
	}
}
