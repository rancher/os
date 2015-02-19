// +build windows

package serial

import (
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
	"unsafe"
)

type serialPort struct {
	f  *os.File
	fd syscall.Handle
	rl sync.Mutex
	wl sync.Mutex
	ro *syscall.Overlapped
	wo *syscall.Overlapped
}

type structDCB struct {
	DCBlength, BaudRate                            uint32
	flags                                          [4]byte
	wReserved, XonLim, XoffLim                     uint16
	ByteSize, Parity, StopBits                     byte
	XonChar, XoffChar, ErrorChar, EofChar, EvtChar byte
	wReserved1                                     uint16
}

type structTimeouts struct {
	ReadIntervalTimeout         uint32
	ReadTotalTimeoutMultiplier  uint32
	ReadTotalTimeoutConstant    uint32
	WriteTotalTimeoutMultiplier uint32
	WriteTotalTimeoutConstant   uint32
}

func openPort(name string, baud int) (rwc io.ReadWriteCloser, err error) {
	if len(name) > 0 && name[0] != '\\' {
		name = "\\\\.\\" + name
	}

	h, err := syscall.CreateFile(syscall.StringToUTF16Ptr(name),
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		0,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_ATTRIBUTE_NORMAL|syscall.FILE_FLAG_OVERLAPPED,
		0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(h), name)
	defer func() {
		if err != nil {
			f.Close()
		}
	}()

	if err = setCommState(h, baud); err != nil {
		return
	}
	if err = setupComm(h, 64, 64); err != nil {
		return
	}
	if err = setCommTimeouts(h); err != nil {
		return
	}
	if err = setCommMask(h); err != nil {
		return
	}

	ro, err := newOverlapped()
	if err != nil {
		return
	}
	wo, err := newOverlapped()
	if err != nil {
		return
	}
	port := new(serialPort)
	port.f = f
	port.fd = h
	port.ro = ro
	port.wo = wo

	return port, nil
}

func (p *serialPort) Close() error {
	return p.f.Close()
}

func (p *serialPort) Write(buf []byte) (int, error) {
	p.wl.Lock()
	defer p.wl.Unlock()

	if err := resetEvent(p.wo.HEvent); err != nil {
		return 0, err
	}
	var n uint32
	err := syscall.WriteFile(p.fd, buf, &n, p.wo)
	if err != nil && err != syscall.ERROR_IO_PENDING {
		return int(n), err
	}
	return getOverlappedResult(p.fd, p.wo)
}

func (p *serialPort) Read(buf []byte) (int, error) {
	if p == nil || p.f == nil {
		return 0, fmt.Errorf("Invalid port on read %v %v", p, p.f)
	}

	p.rl.Lock()
	defer p.rl.Unlock()

	if err := resetEvent(p.ro.HEvent); err != nil {
		return 0, err
	}
	var done uint32
	err := syscall.ReadFile(p.fd, buf, &done, p.ro)
	if err != nil && err != syscall.ERROR_IO_PENDING {
		return int(done), err
	}
	return getOverlappedResult(p.fd, p.ro)
}

var (
	nSetCommState,
	nSetCommTimeouts,
	nSetCommMask,
	nSetupComm,
	nGetOverlappedResult,
	nCreateEvent,
	nResetEvent uintptr
)

func init() {
	k32, err := syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		panic("LoadLibrary " + err.Error())
	}
	defer syscall.FreeLibrary(k32)

	nSetCommState = getProcAddr(k32, "SetCommState")
	nSetCommTimeouts = getProcAddr(k32, "SetCommTimeouts")
	nSetCommMask = getProcAddr(k32, "SetCommMask")
	nSetupComm = getProcAddr(k32, "SetupComm")
	nGetOverlappedResult = getProcAddr(k32, "GetOverlappedResult")
	nCreateEvent = getProcAddr(k32, "CreateEventW")
	nResetEvent = getProcAddr(k32, "ResetEvent")
}

func getProcAddr(lib syscall.Handle, name string) uintptr {
	addr, err := syscall.GetProcAddress(lib, name)
	if err != nil {
		panic(name + " " + err.Error())
	}
	return addr
}

func setCommState(h syscall.Handle, baud int) error {
	var params structDCB
	params.DCBlength = uint32(unsafe.Sizeof(params))

	params.flags[0] = 0x01  // fBinary
	params.flags[0] |= 0x10 // Assert DSR

	params.BaudRate = uint32(baud)
	params.ByteSize = 8

	r, _, err := syscall.Syscall(nSetCommState, 2, uintptr(h), uintptr(unsafe.Pointer(&params)), 0)
	if r == 0 {
		return err
	}
	return nil
}

func setCommTimeouts(h syscall.Handle) error {
	var timeouts structTimeouts
	const MAXDWORD = 1<<32 - 1
	timeouts.ReadIntervalTimeout = MAXDWORD
	timeouts.ReadTotalTimeoutMultiplier = MAXDWORD
	timeouts.ReadTotalTimeoutConstant = MAXDWORD - 1

	/* From http://msdn.microsoft.com/en-us/library/aa363190(v=VS.85).aspx

		 For blocking I/O see below:

		 Remarks:

		 If an application sets ReadIntervalTimeout and
		 ReadTotalTimeoutMultiplier to MAXDWORD and sets
		 ReadTotalTimeoutConstant to a value greater than zero and
		 less than MAXDWORD, one of the following occurs when the
		 ReadFile function is called:

		 If there are any bytes in the input buffer, ReadFile returns
		       immediately with the bytes in the buffer.

		 If there are no bytes in the input buffer, ReadFile waits
	               until a byte arrives and then returns immediately.

		 If no bytes arrive within the time specified by
		       ReadTotalTimeoutConstant, ReadFile times out.
	*/

	r, _, err := syscall.Syscall(nSetCommTimeouts, 2, uintptr(h), uintptr(unsafe.Pointer(&timeouts)), 0)
	if r == 0 {
		return err
	}
	return nil
}

func setupComm(h syscall.Handle, in, out int) error {
	r, _, err := syscall.Syscall(nSetupComm, 3, uintptr(h), uintptr(in), uintptr(out))
	if r == 0 {
		return err
	}
	return nil
}

func setCommMask(h syscall.Handle) error {
	const EV_RXCHAR = 0x0001
	r, _, err := syscall.Syscall(nSetCommMask, 2, uintptr(h), EV_RXCHAR, 0)
	if r == 0 {
		return err
	}
	return nil
}

func resetEvent(h syscall.Handle) error {
	r, _, err := syscall.Syscall(nResetEvent, 1, uintptr(h), 0, 0)
	if r == 0 {
		return err
	}
	return nil
}

func newOverlapped() (*syscall.Overlapped, error) {
	var overlapped syscall.Overlapped
	r, _, err := syscall.Syscall6(nCreateEvent, 4, 0, 1, 0, 0, 0, 0)
	if r == 0 {
		return nil, err
	}
	overlapped.HEvent = syscall.Handle(r)
	return &overlapped, nil
}

func getOverlappedResult(h syscall.Handle, overlapped *syscall.Overlapped) (int, error) {
	var n int
	r, _, err := syscall.Syscall6(nGetOverlappedResult, 4,
		uintptr(h),
		uintptr(unsafe.Pointer(overlapped)),
		uintptr(unsafe.Pointer(&n)), 1, 0, 0)
	if r == 0 {
		return n, err
	}

	return n, nil
}
