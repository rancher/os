// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Copied and simplified from ../../x86/x86asm/objdumpext_test.go.

package armasm

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
)

const objdumpPath = "/usr/local/bin/arm-linux-elf-objdump"

func testObjdumpARM(t *testing.T, generate func(func([]byte))) {
	testObjdumpArch(t, generate, ModeARM)
}

func testObjdumpArch(t *testing.T, generate func(func([]byte)), arch Mode) {
	if testing.Short() {
		t.Skip("skipping objdump test in short mode")
	}
	if _, err := os.Stat(objdumpPath); err != nil {
		t.Skip(err)
	}

	testExtDis(t, "gnu", arch, objdump, generate, allowedMismatchObjdump)
}

func objdump(ext *ExtDis) error {
	// File already written with instructions; add ELF header.
	if ext.Arch == ModeARM {
		if err := writeELF32(ext.File, ext.Size); err != nil {
			return err
		}
	} else {
		panic("unknown arch")
	}

	b, err := ext.Run(objdumpPath, "-d", "-z", ext.File.Name())
	if err != nil {
		return err
	}

	var (
		nmatch  int
		reading bool
		next    uint32 = start
		addr    uint32
		encbuf  [4]byte
		enc     []byte
		text    string
	)
	flush := func() {
		if addr == next {
			if m := pcrel.FindStringSubmatch(text); m != nil {
				targ, _ := strconv.ParseUint(m[2], 16, 64)
				text = fmt.Sprintf("%s .%+#x", m[1], int32(uint32(targ)-addr-uint32(len(enc))))
			}
			if strings.HasPrefix(text, "stmia") {
				text = "stm" + text[5:]
			}
			if strings.HasPrefix(text, "stmfd") {
				text = "stmdb" + text[5:]
			}
			if strings.HasPrefix(text, "ldmfd") {
				text = "ldm" + text[5:]
			}
			text = strings.Replace(text, "#0.0", "#0", -1)
			if text == "undefined" && len(enc) == 4 {
				text = "error: unknown instruction"
				enc = nil
			}
			if len(enc) == 4 {
				// prints as word but we want to record bytes
				enc[0], enc[3] = enc[3], enc[0]
				enc[1], enc[2] = enc[2], enc[1]
			}
			ext.Dec <- ExtInst{addr, encbuf, len(enc), text}
			encbuf = [4]byte{}
			enc = nil
			next += 4
		}
	}
	var textangle = []byte("<.text>:")
	for {
		line, err := b.ReadSlice('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("reading objdump output: %v", err)
		}
		if bytes.Contains(line, textangle) {
			reading = true
			continue
		}
		if !reading {
			continue
		}
		if debug {
			os.Stdout.Write(line)
		}
		if enc1 := parseContinuation(line, encbuf[:len(enc)]); enc1 != nil {
			enc = enc1
			continue
		}
		flush()
		nmatch++
		addr, enc, text = parseLine(line, encbuf[:0])
		if addr > next {
			return fmt.Errorf("address out of sync expected <= %#x at %q in:\n%s", next, line, line)
		}
	}
	flush()
	if next != start+uint32(ext.Size) {
		return fmt.Errorf("not enough results found [%d %d]", next, start+ext.Size)
	}
	if err := ext.Wait(); err != nil {
		return fmt.Errorf("exec: %v", err)
	}

	return nil
}

var (
	undefined      = []byte("<UNDEFINED>")
	unpredictable  = []byte("<UNPREDICTABLE>")
	illegalShifter = []byte("<illegal shifter operand>")
)

func parseLine(line []byte, encstart []byte) (addr uint32, enc []byte, text string) {
	oline := line
	i := index(line, ":\t")
	if i < 0 {
		log.Fatalf("cannot parse disassembly: %q", oline)
	}
	x, err := strconv.ParseUint(string(trimSpace(line[:i])), 16, 32)
	if err != nil {
		log.Fatalf("cannot parse disassembly: %q", oline)
	}
	addr = uint32(x)
	line = line[i+2:]
	i = bytes.IndexByte(line, '\t')
	if i < 0 {
		log.Fatalf("cannot parse disassembly: %q", oline)
	}
	enc, ok := parseHex(line[:i], encstart)
	if !ok {
		log.Fatalf("cannot parse disassembly: %q", oline)
	}
	line = trimSpace(line[i:])
	if bytes.Contains(line, undefined) {
		text = "undefined"
		return
	}
	if bytes.Contains(line, illegalShifter) {
		text = "undefined"
		return
	}
	if false && bytes.Contains(line, unpredictable) {
		text = "unpredictable"
		return
	}
	if i := bytes.IndexByte(line, ';'); i >= 0 {
		line = trimSpace(line[:i])
	}
	text = string(fixSpace(line))
	return
}

func parseContinuation(line []byte, enc []byte) []byte {
	i := index(line, ":\t")
	if i < 0 {
		return nil
	}
	line = line[i+1:]
	enc, _ = parseHex(line, enc)
	return enc
}

// writeELF32 writes an ELF32 header to the file,
// describing a text segment that starts at start
// and extends for size bytes.
func writeELF32(f *os.File, size int) error {
	f.Seek(0, 0)
	var hdr elf.Header32
	var prog elf.Prog32
	var sect elf.Section32
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, &hdr)
	off1 := buf.Len()
	binary.Write(&buf, binary.LittleEndian, &prog)
	off2 := buf.Len()
	binary.Write(&buf, binary.LittleEndian, &sect)
	off3 := buf.Len()
	buf.Reset()
	data := byte(elf.ELFDATA2LSB)
	hdr = elf.Header32{
		Ident:     [16]byte{0x7F, 'E', 'L', 'F', 1, data, 1},
		Type:      2,
		Machine:   uint16(elf.EM_ARM),
		Version:   1,
		Entry:     start,
		Phoff:     uint32(off1),
		Shoff:     uint32(off2),
		Flags:     0x05000002,
		Ehsize:    uint16(off1),
		Phentsize: uint16(off2 - off1),
		Phnum:     1,
		Shentsize: uint16(off3 - off2),
		Shnum:     3,
		Shstrndx:  2,
	}
	binary.Write(&buf, binary.LittleEndian, &hdr)
	prog = elf.Prog32{
		Type:   1,
		Off:    start,
		Vaddr:  start,
		Paddr:  start,
		Filesz: uint32(size),
		Memsz:  uint32(size),
		Flags:  5,
		Align:  start,
	}
	binary.Write(&buf, binary.LittleEndian, &prog)
	binary.Write(&buf, binary.LittleEndian, &sect) // NULL section
	sect = elf.Section32{
		Name:      1,
		Type:      uint32(elf.SHT_PROGBITS),
		Addr:      start,
		Off:       start,
		Size:      uint32(size),
		Flags:     uint32(elf.SHF_ALLOC | elf.SHF_EXECINSTR),
		Addralign: 4,
	}
	binary.Write(&buf, binary.LittleEndian, &sect) // .text
	sect = elf.Section32{
		Name:      uint32(len("\x00.text\x00")),
		Type:      uint32(elf.SHT_STRTAB),
		Addr:      0,
		Off:       uint32(off2 + (off3-off2)*3),
		Size:      uint32(len("\x00.text\x00.shstrtab\x00")),
		Addralign: 1,
	}
	binary.Write(&buf, binary.LittleEndian, &sect)
	buf.WriteString("\x00.text\x00.shstrtab\x00")
	f.Write(buf.Bytes())
	return nil
}
