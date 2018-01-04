// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Parsing of Go intermediate object files and archives.

package objfile

/*


import (
	"debug/dwarf"
	"debug/gosym"
	"errors"
	"fmt"
	"os"

	"github.com/google/gops/internal/sys"

	"github.com/google/gops/internal/goobj"
)

type goobjFile struct {
	goobj *goobj.Package
	f     *os.File // the underlying .o or .a file
}

func openGoobj(r *os.File) (rawFile, error) {
	f, err := goobj.Parse(r, `""`)
	if err != nil {
		return nil, err
	}
	return &goobjFile{goobj: f, f: r}, nil
}

func goobjName(id goobj.SymID) string {
	if id.Version == 0 {
		return id.Name
	}
	return fmt.Sprintf("%s<%d>", id.Name, id.Version)
}

func (f *goobjFile) symbols() ([]Sym, error) {
	seen := make(map[goobj.SymID]bool)

	var syms []Sym
	for _, s := range f.goobj.Syms {
		seen[s.SymID] = true
		sym := Sym{Addr: uint64(s.Data.Offset), Name: goobjName(s.SymID), Size: int64(s.Size), Type: s.Type.Name, Code: '?'}
		switch s.Kind {
		case goobj.STEXT, goobj.SELFRXSECT:
			sym.Code = 'T'
		case goobj.STYPE, goobj.SSTRING, goobj.SGOSTRING, goobj.SGOFUNC, goobj.SRODATA, goobj.SFUNCTAB, goobj.STYPELINK, goobj.SITABLINK, goobj.SSYMTAB, goobj.SPCLNTAB, goobj.SELFROSECT:
			sym.Code = 'R'
		case goobj.SMACHOPLT, goobj.SELFSECT, goobj.SMACHO, goobj.SMACHOGOT, goobj.SNOPTRDATA, goobj.SINITARR, goobj.SDATA, goobj.SWINDOWS:
			sym.Code = 'D'
		case goobj.SBSS, goobj.SNOPTRBSS, goobj.STLSBSS:
			sym.Code = 'B'
		case goobj.SXREF, goobj.SMACHOSYMSTR, goobj.SMACHOSYMTAB, goobj.SMACHOINDIRECTPLT, goobj.SMACHOINDIRECTGOT, goobj.SFILE, goobj.SFILEPATH, goobj.SCONST, goobj.SDYNIMPORT, goobj.SHOSTOBJ:
			sym.Code = 'X' // should not see
		}
		if s.Version != 0 {
			sym.Code += 'a' - 'A'
		}
		for i, r := range s.Reloc {
			sym.Relocs = append(sym.Relocs, Reloc{Addr: uint64(s.Data.Offset) + uint64(r.Offset), Size: uint64(r.Size), Stringer: &s.Reloc[i]})
		}
		syms = append(syms, sym)
	}

	for _, s := range f.goobj.Syms {
		for _, r := range s.Reloc {
			if !seen[r.Sym] {
				seen[r.Sym] = true
				sym := Sym{Name: goobjName(r.Sym), Code: 'U'}
				if s.Version != 0 {
					// should not happen but handle anyway
					sym.Code = 'u'
				}
				syms = append(syms, sym)
			}
		}
	}

	return syms, nil
}

func (f *goobjFile) pcln() (textStart uint64, symtab, pclntab []byte, err error) {
	// Should never be called.  We implement Liner below, callers
	// should use that instead.
	return 0, nil, nil, fmt.Errorf("pcln not available in go object file")
}

// Find returns the file name, line, and function data for the given pc.
// Returns "",0,nil if unknown.
// This function implements the Liner interface in preference to pcln() above.
func (f *goobjFile) PCToLine(pc uint64) (string, int, *gosym.Func) {
	// TODO: this is really inefficient.  Binary search?  Memoize last result?
	var arch *sys.Arch
	for _, a := range sys.Archs {
		if a.Name == f.goobj.Arch {
			arch = a
			break
		}
	}
	if arch == nil {
		return "", 0, nil
	}
	for _, s := range f.goobj.Syms {
		if pc < uint64(s.Data.Offset) || pc >= uint64(s.Data.Offset+s.Data.Size) {
			continue
		}
		if s.Func == nil {
			return "", 0, nil
		}
		pcfile := make([]byte, s.Func.PCFile.Size)
		_, err := f.f.ReadAt(pcfile, s.Func.PCFile.Offset)
		if err != nil {
			return "", 0, nil
		}
		fileID := gosym.PCValue(pcfile, pc-uint64(s.Data.Offset), arch.MinLC)
		fileName := s.Func.File[fileID]
		pcline := make([]byte, s.Func.PCLine.Size)
		_, err = f.f.ReadAt(pcline, s.Func.PCLine.Offset)
		if err != nil {
			return "", 0, nil
		}
		line := gosym.PCValue(pcline, pc-uint64(s.Data.Offset), arch.MinLC)
		// Note: we provide only the name in the Func structure.
		// We could provide more if needed.
		return fileName, line, &gosym.Func{Sym: &gosym.Sym{Name: s.Name}}
	}
	return "", 0, nil
}

// We treat the whole object file as the text section.
func (f *goobjFile) text() (textStart uint64, text []byte, err error) {
	var info os.FileInfo
	info, err = f.f.Stat()
	if err != nil {
		return
	}
	text = make([]byte, info.Size())
	_, err = f.f.ReadAt(text, 0)
	return
}

func (f *goobjFile) goarch() string {
	return f.goobj.Arch
}

func (f *goobjFile) loadAddress() (uint64, error) {
	return 0, fmt.Errorf("unknown load address")
}

func (f *goobjFile) dwarf() (*dwarf.Data, error) {
	return nil, errors.New("no DWARF data in go object file")
}
*/
