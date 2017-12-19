/*********************************************************
 * Copyright (C) 1998-2015 VMware, Inc. All rights reserved.
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms of the GNU Lesser General Public License as published
 * by the Free Software Foundation version 2.1 and no later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY
 * or FITNESS FOR A PARTICULAR PURPOSE.  See the Lesser GNU General Public
 * License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with this program; if not, write to the Free Software Foundation, Inc.,
 * 51 Franklin St, Fifth Floor, Boston, MA  02110-1301 USA.
 *
 *********************************************************/

/*********************************************************
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 *
 * 1. Redistributions of source code must retain the above copyright
 *    notice, this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE REGENTS AND CONTRIBUTORS ``AS IS'' AND
 * ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED.  IN NO EVENT SHALL THE REGENTS OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS
 * OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
 * HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
 * LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
 * OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
 * SUCH DAMAGE.
 *
 *********************************************************/

/*********************************************************
 * The contents of this file are subject to the terms of the Common
 * Development and Distribution License (the "License") version 1.0
 * and no later version.  You may not use this file except in
 * compliance with the License.
 *
 * You can obtain a copy of the License at
 *         http://www.opensource.org/licenses/cddl1.php
 *
 * See the License for the specific language governing permissions
 * and limitations under the License.
 *
 *********************************************************/

/*
 * vm_assert.h --
 *
 *	The basic assertion facility for all VMware code.
 *
 *      For proper use, see bora/doc/assert and
 *      http://vmweb.vmware.com/~mts/WebSite/guide/programming/asserts.html.
 */

#ifndef _VM_ASSERT_H_
#define _VM_ASSERT_H_

// XXX not necessary except some places include vm_assert.h improperly
#include "vm_basic_types.h"

#ifdef __cplusplus
extern "C" {
#endif

/*
 * Some bits of vmcore are used in VMKernel code and cannot have
 * the VMKERNEL define due to other header dependencies.
 */
#if defined(VMKERNEL) && !defined(VMKPANIC)
#define VMKPANIC 1
#endif

/*
 * Internal macros, functions, and strings
 *
 * The monitor wants to save space at call sites, so it has specialized
 * functions for each situation.  User level wants to save on implementation
 * so it uses generic functions.
 */

#if !defined VMM || defined MONITOR_APP // {

#if defined (VMKPANIC) 
#include "vmk_assert.h"
#else /* !VMKPANIC */
#define _ASSERT_PANIC(name) \
           Panic(_##name##Fmt "\n", __FILE__, __LINE__)
#define _ASSERT_PANIC_BUG(bug, name) \
           Panic(_##name##Fmt " bugNr=%d\n", __FILE__, __LINE__, bug)
#define _ASSERT_PANIC_NORETURN(name) \
           Panic(_##name##Fmt "\n", __FILE__, __LINE__)
#define _ASSERT_PANIC_BUG_NORETURN(bug, name) \
           Panic(_##name##Fmt " bugNr=%d\n", __FILE__, __LINE__, bug)
#endif /* VMKPANIC */

#endif // }


// These strings don't have newline so that a bug can be tacked on.
#define _AssertPanicFmt            "PANIC %s:%d"
#define _AssertAssertFmt           "ASSERT %s:%d"
#define _AssertVerifyFmt           "VERIFY %s:%d"
#define _AssertNotImplementedFmt   "NOT_IMPLEMENTED %s:%d"
#define _AssertNotReachedFmt       "NOT_REACHED %s:%d"
#define _AssertMemAllocFmt         "MEM_ALLOC %s:%d"
#define _AssertNotTestedFmt        "NOT_TESTED %s:%d"


/*
 * Panic and log functions
 */

void Log(const char *fmt, ...) PRINTF_DECL(1, 2);
void Warning(const char *fmt, ...) PRINTF_DECL(1, 2);
#if defined VMKPANIC
void Panic_SaveRegs(void);

#ifdef VMX86_DEBUG
void Panic_NoSave(const char *fmt, ...) PRINTF_DECL(1, 2);
#else
NORETURN void Panic_NoSave(const char *fmt, ...) PRINTF_DECL(1, 2);
#endif

NORETURN void Panic_NoSaveNoReturn(const char *fmt, ...) PRINTF_DECL(1, 2);

#define Panic(fmt...) do { \
   Panic_SaveRegs();       \
   Panic_NoSave(fmt);      \
} while(0)

#define Panic_NoReturn(fmt...) do { \
   Panic_SaveRegs();                \
   Panic_NoSaveNoReturn(fmt);       \
} while(0)

#else
NORETURN void Panic(const char *fmt, ...) PRINTF_DECL(1, 2);
#endif

void LogThrottled(uint32 *count, const char *fmt, ...) PRINTF_DECL(2, 3);
void WarningThrottled(uint32 *count, const char *fmt, ...) PRINTF_DECL(2, 3);


#ifndef ASSERT_IFNOT
   /*
    * PR 271512: When compiling with gcc, catch assignments inside an ASSERT.
    *
    * 'UNLIKELY' is defined with __builtin_expect, which does not warn when
    * passed an assignment (gcc bug 36050). To get around this, we put 'cond'
    * in an 'if' statement and make sure it never gets executed by putting
    * that inside of 'if (0)'. We use gcc's statement expression syntax to
    * make ASSERT an expression because some code uses it that way.
    *
    * Since statement expression syntax is a gcc extension and since it's
    * not clear if this is a problem with other compilers, the ASSERT
    * definition was not changed for them. Using a bare 'cond' with the
    * ternary operator may provide a solution.
    */

   #ifdef __GNUC__
      #define ASSERT_IFNOT(cond, panic)                                       \
         ({if (UNLIKELY(!(cond))) { panic; if (0) { if (cond) {;}}} (void)0;})
   #else
      #define ASSERT_IFNOT(cond, panic)                                       \
         (UNLIKELY(!(cond)) ? (panic) : (void)0)
   #endif
#endif


/*
 * Assert, panic, and log macros
 *
 * Some of these are redefined below undef !VMX86_DEBUG.
 * ASSERT() is special cased because of interaction with Windows DDK.
 */

#if defined VMX86_DEBUG
#undef  ASSERT
#define ASSERT(cond) ASSERT_IFNOT(cond, _ASSERT_PANIC(AssertAssert))
#define ASSERT_BUG(bug, cond) \
           ASSERT_IFNOT(cond, _ASSERT_PANIC_BUG(bug, AssertAssert))
#endif

#undef  VERIFY
#define VERIFY(cond) \
           ASSERT_IFNOT(cond, _ASSERT_PANIC_NORETURN(AssertVerify))
#define VERIFY_BUG(bug, cond) \
           ASSERT_IFNOT(cond, _ASSERT_PANIC_BUG_NORETURN(bug, AssertVerify))

#define PANIC()        _ASSERT_PANIC(AssertPanic)
#define PANIC_BUG(bug) _ASSERT_PANIC_BUG(bug, AssertPanic)

#define ASSERT_NOT_IMPLEMENTED(cond) \
           ASSERT_IFNOT(cond, NOT_IMPLEMENTED())
#define ASSERT_NOT_IMPLEMENTED_BUG(bug, cond) \
           ASSERT_IFNOT(cond, NOT_IMPLEMENTED_BUG(bug))

#if defined VMKPANIC || defined VMM
#define NOT_IMPLEMENTED()        _ASSERT_PANIC_NORETURN(AssertNotImplemented)
#else
#define NOT_IMPLEMENTED()        _ASSERT_PANIC(AssertNotImplemented)
#endif

#if defined VMM
#define NOT_IMPLEMENTED_BUG(bug) \
          _ASSERT_PANIC_BUG_NORETURN(bug, AssertNotImplemented)
#else 
#define NOT_IMPLEMENTED_BUG(bug) _ASSERT_PANIC_BUG(bug, AssertNotImplemented)
#endif

#if defined VMKPANIC || defined VMM
#define NOT_REACHED()            _ASSERT_PANIC_NORETURN(AssertNotReached)
#else
#define NOT_REACHED()            _ASSERT_PANIC(AssertNotReached)
#endif

#define ASSERT_MEM_ALLOC(cond) \
           ASSERT_IFNOT(cond, _ASSERT_PANIC(AssertMemAlloc))

#ifdef VMX86_DEVEL
#define ASSERT_DEVEL(cond) ASSERT(cond)
#define NOT_TESTED()       Warning(_AssertNotTestedFmt "\n", __FILE__, __LINE__)
#else
#define ASSERT_DEVEL(cond) ((void)0)
#define NOT_TESTED()       Log(_AssertNotTestedFmt "\n", __FILE__, __LINE__)
#endif

#define ASSERT_NO_INTERRUPTS()  ASSERT(!INTERRUPTS_ENABLED())
#define ASSERT_HAS_INTERRUPTS() ASSERT(INTERRUPTS_ENABLED())

#define ASSERT_NOT_TESTED(cond) (UNLIKELY(!(cond)) ? NOT_TESTED() : (void)0)
#define NOT_TESTED_ONCE()       DO_ONCE(NOT_TESTED())

#define NOT_TESTED_1024()                                               \
   do {                                                                 \
      static uint16 count = 0;                                          \
      if (UNLIKELY(count == 0)) { NOT_TESTED(); }                       \
      count = (count + 1) & 1023;                                       \
   } while (0)

#define LOG_ONCE(_s) DO_ONCE(Log _s)


/*
 * Redefine macros that are only in debug versions
 */

#if !defined VMX86_DEBUG // {

#undef  ASSERT
#define ASSERT(cond)          ((void)0)
#define ASSERT_BUG(bug, cond) ((void)0)

/*
 * Expand NOT_REACHED() as appropriate for each situation.
 *
 * Mainly, we want the compiler to infer the same control-flow
 * information as it would from Panic().  Otherwise, different
 * compilation options will lead to different control-flow-derived
 * errors, causing some make targets to fail while others succeed.
 *
 * VC++ has the __assume() built-in function which we don't trust
 * (see bug 43485); gcc has no such construct; we just panic in
 * userlevel code.  The monitor doesn't want to pay the size penalty
 * (measured at 212 bytes for the release vmm for a minimal infinite
 * loop; panic would cost even more) so it does without and lives
 * with the inconsistency.
 */

#if defined VMKPANIC || defined VMM
#undef  NOT_REACHED
#if defined __GNUC__ && (__GNUC__ > 4 || __GNUC__ == 4 && __GNUC_MINOR__ >= 5)
#define NOT_REACHED() (__builtin_unreachable())
#else
#define NOT_REACHED() ((void)0)
#endif
#else
// keep debug definition
#endif

#undef LOG_UNEXPECTED
#define LOG_UNEXPECTED(bug)     ((void)0)

#undef  ASSERT_NOT_TESTED
#define ASSERT_NOT_TESTED(cond) ((void)0)
#undef  NOT_TESTED
#define NOT_TESTED()            ((void)0)
#undef  NOT_TESTED_ONCE
#define NOT_TESTED_ONCE()       ((void)0)
#undef  NOT_TESTED_1024
#define NOT_TESTED_1024()       ((void)0)

#endif // !VMX86_DEBUG }


/*
 * Compile-time assertions.
 *
 * ASSERT_ON_COMPILE does not use the common
 * switch (0) { case 0: case (e): ; } trick because some compilers (e.g. MSVC)
 * generate code for it.
 *
 * The implementation uses both enum and typedef because the typedef alone is
 * insufficient; gcc allows arrays to be declared with non-constant expressions
 * (even in typedefs, where it makes no sense).
 *
 * NOTE: if GCC ever changes so that it ignores unused types altogether, this
 * assert might not fire!  We explicitly mark it as unused because GCC 4.8+
 * uses -Wunused-local-typedefs as part of -Wall, which means the typedef will
 * generate a warning.
 */

#if defined(_Static_assert) || defined(__cplusplus) ||                         \
    !defined(__GNUC__) || __GNUC__ < 4 || (__GNUC__ == 4 && __GNUC_MINOR__ < 6)
#define ASSERT_ON_COMPILE(e) \
   do { \
      enum { AssertOnCompileMisused = ((e) ? 1 : -1) }; \
      UNUSED_TYPE(typedef char AssertOnCompileFailed[AssertOnCompileMisused]); \
   } while (0)
#else
#define ASSERT_ON_COMPILE(e) \
   do {                      \
      _Static_assert(e, #e); \
   } while (0);
#endif

/*
 * To put an ASSERT_ON_COMPILE() outside a function, wrap it
 * in MY_ASSERTS().  The first parameter must be unique in
 * each .c file where it appears.  For example,
 *
 * MY_ASSERTS(FS3_INT,
 *    ASSERT_ON_COMPILE(sizeof(FS3_DiskLock) == 128);
 *    ASSERT_ON_COMPILE(sizeof(FS3_DiskLockReserved) == DISK_BLOCK_SIZE);
 *    ASSERT_ON_COMPILE(sizeof(FS3_DiskBlock) == DISK_BLOCK_SIZE);
 *    ASSERT_ON_COMPILE(sizeof(Hardware_DMIUUID) == 16);
 * )
 *
 * Caution: ASSERT() within MY_ASSERTS() is silently ignored.
 * The same goes for anything else not evaluated at compile time.
 */

#define MY_ASSERTS(name, assertions) \
   static INLINE void name(void) {   \
      assertions                     \
   }

#ifdef __cplusplus
} /* extern "C" */
#endif

#endif /* ifndef _VM_ASSERT_H_ */
