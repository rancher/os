// Inferno utils/6l/pass.c
// https://bitbucket.org/inferno-os/inferno-os/src/default/utils/6l/pass.c
//
//	Copyright © 1994-1999 Lucent Technologies Inc.  All rights reserved.
//	Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
//	Portions Copyright © 1997-1999 Vita Nuova Limited
//	Portions Copyright © 2000-2007 Vita Nuova Holdings Limited (www.vitanuova.com)
//	Portions Copyright © 2004,2006 Bruce Ellis
//	Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
//	Revisions Copyright © 2000-2007 Lucent Technologies Inc. and others
//	Portions Copyright © 2009 The Go Authors. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package obj

// Code and data passes.

func Brchain(ctxt *Link, p *Prog) *Prog {
	for i := 0; i < 20; i++ {
		if p == nil || p.As != AJMP || p.Pcond == nil {
			return p
		}
		p = p.Pcond
	}

	return nil
}

func brloop(ctxt *Link, p *Prog) *Prog {
	var q *Prog

	c := 0
	for q = p; q != nil; q = q.Pcond {
		if q.As != AJMP || q.Pcond == nil {
			break
		}
		c++
		if c >= 5000 {
			return nil
		}
	}

	return q
}

func checkaddr(ctxt *Link, p *Prog, a *Addr) {
	// Check expected encoding, especially TYPE_CONST vs TYPE_ADDR.
	switch a.Type {
	case TYPE_NONE:
		return

	case TYPE_BRANCH:
		if a.Reg != 0 || a.Index != 0 || a.Scale != 0 || a.Name != 0 {
			break
		}
		return

	case TYPE_TEXTSIZE:
		if a.Reg != 0 || a.Index != 0 || a.Scale != 0 || a.Name != 0 {
			break
		}
		return

		//if(a->u.bits != 0)
	//	break;
	case TYPE_MEM:
		return

		// TODO(rsc): After fixing SHRQ, check a->index != 0 too.
	case TYPE_CONST:
		if a.Name != 0 || a.Sym != nil || a.Reg != 0 {
			ctxt.Diag("argument is TYPE_CONST, should be TYPE_ADDR, in %v", p)
			return
		}

		if a.Reg != 0 || a.Scale != 0 || a.Name != 0 || a.Sym != nil || a.Val != nil {
			break
		}
		return

	case TYPE_FCONST, TYPE_SCONST:
		if a.Reg != 0 || a.Index != 0 || a.Scale != 0 || a.Name != 0 || a.Offset != 0 || a.Sym != nil {
			break
		}
		return

	// TODO(rsc): After fixing PINSRQ, check a->offset != 0 too.
	// TODO(rsc): After fixing SHRQ, check a->index != 0 too.
	case TYPE_REG:
		if a.Scale != 0 || a.Name != 0 || a.Sym != nil {
			break
		}
		return

	case TYPE_ADDR:
		if a.Val != nil {
			break
		}
		if a.Reg == 0 && a.Index == 0 && a.Scale == 0 && a.Name == 0 && a.Sym == nil {
			ctxt.Diag("argument is TYPE_ADDR, should be TYPE_CONST, in %v", p)
		}
		return

	case TYPE_SHIFT:
		if a.Index != 0 || a.Scale != 0 || a.Name != 0 || a.Sym != nil || a.Val != nil {
			break
		}
		return

	case TYPE_REGREG:
		if a.Index != 0 || a.Scale != 0 || a.Name != 0 || a.Sym != nil || a.Val != nil {
			break
		}
		return

	case TYPE_REGREG2:
		return

	case TYPE_REGLIST:
		return

	// Expect sym and name to be set, nothing else.
	// Technically more is allowed, but this is only used for *name(SB).
	case TYPE_INDIR:
		if a.Reg != 0 || a.Index != 0 || a.Scale != 0 || a.Name == 0 || a.Offset != 0 || a.Sym == nil || a.Val != nil {
			break
		}
		return
	}

	ctxt.Diag("invalid encoding for argument %v", p)
}

func linkpatch(ctxt *Link, sym *LSym) {
	var c int32
	var name string
	var q *Prog

	ctxt.Cursym = sym

	for p := sym.Text; p != nil; p = p.Link {
		checkaddr(ctxt, p, &p.From)
		if p.From3 != nil {
			checkaddr(ctxt, p, p.From3)
		}
		checkaddr(ctxt, p, &p.To)

		if ctxt.Arch.Progedit != nil {
			ctxt.Arch.Progedit(ctxt, p)
		}
		if p.To.Type != TYPE_BRANCH {
			continue
		}
		if p.To.Val != nil {
			// TODO: Remove To.Val.(*Prog) in favor of p->pcond.
			p.Pcond = p.To.Val.(*Prog)
			continue
		}

		if p.To.Sym != nil {
			continue
		}
		c = int32(p.To.Offset)
		for q = sym.Text; q != nil; {
			if int64(c) == q.Pc {
				break
			}
			if q.Forwd != nil && int64(c) >= q.Forwd.Pc {
				q = q.Forwd
			} else {
				q = q.Link
			}
		}

		if q == nil {
			name = "<nil>"
			if p.To.Sym != nil {
				name = p.To.Sym.Name
			}
			ctxt.Diag("branch out of range (%#x)\n%v [%s]", uint32(c), p, name)
			p.To.Type = TYPE_NONE
		}

		p.To.Val = q
		p.Pcond = q
	}

	if ctxt.Flag_optimize {
		for p := sym.Text; p != nil; p = p.Link {
			if p.Pcond != nil {
				p.Pcond = brloop(ctxt, p.Pcond)
				if p.Pcond != nil {
					if p.To.Type == TYPE_BRANCH {
						p.To.Offset = p.Pcond.Pc
					}
				}
			}
		}
	}
}
