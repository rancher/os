/*********************************************************
 * Copyright (C) 2005-2015 VMware, Inc. All rights reserved.
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
 * backdoorGcc64.c --
 *
 *      Implements the real work for guest-side backdoor for GCC, 64-bit
 *      target (supports inline ASM, GAS syntax). The asm sections are marked
 *      volatile since vmware can change the registers content without the
 *      compiler knowing it.
 *
 *      See backdoorGCC32.c (from which this code was mostly copied) for
 *      details on why the ASM is written this way. Also note that it might be
 *      possible to write the asm blocks using the symbolic operand specifiers
 *      in such a way that the same asm would generate correct code for both
 *      32-bit and 64-bit targets, but I'm too lazy to figure it all out.
 *      --rrdharan
 */
#ifdef __cplusplus
extern "C" {
#endif

#include "backdoor.h"
#include "backdoorInt.h"


/*
 *----------------------------------------------------------------------------
 *
 * Backdoor_InOut --
 *
 *      Send a low-bandwidth basic request (16 bytes) to vmware, and return its
 *      reply (24 bytes).
 *
 * Results:
 *      Host-side response returned in bp IN/OUT parameter.
 *
 * Side effects:
 *      Pokes the backdoor.
 *
 *----------------------------------------------------------------------------
 */

void
Backdoor_InOut(Backdoor_proto *myBp) // IN/OUT
{
   uint64 dummy;

   __asm__ __volatile__(
#ifdef __APPLE__
        /*
         * Save %rbx on the stack because the Mac OS GCC doesn't want us to
         * clobber it - it erroneously thinks %rbx is the PIC register.
         * (Radar bug 7304232)
         */
        "pushq %%rbx"           "\n\t"
#endif
        "pushq %%rax"           "\n\t"
        "movq 40(%%rax), %%rdi" "\n\t"
        "movq 32(%%rax), %%rsi" "\n\t"
        "movq 24(%%rax), %%rdx" "\n\t"
        "movq 16(%%rax), %%rcx" "\n\t"
        "movq  8(%%rax), %%rbx" "\n\t"
        "movq   (%%rax), %%rax" "\n\t"
        "inl %%dx, %%eax"       "\n\t"  /* NB: There is no inq instruction */
        "xchgq %%rax, (%%rsp)"  "\n\t"
        "movq %%rdi, 40(%%rax)" "\n\t"
        "movq %%rsi, 32(%%rax)" "\n\t"
        "movq %%rdx, 24(%%rax)" "\n\t"
        "movq %%rcx, 16(%%rax)" "\n\t"
        "movq %%rbx,  8(%%rax)" "\n\t"
        "popq          (%%rax)" "\n\t"
#ifdef __APPLE__
        "popq %%rbx"            "\n\t"
#endif
      : "=a" (dummy)
      : "0" (myBp)
      /*
       * vmware can modify the whole VM state without the compiler knowing
       * it. So far it does not modify EFLAGS. --hpreg
       */
      :
#ifndef __APPLE__
      /* %rbx is unchanged at the end of the function on Mac OS. */
      "rbx",
#endif
      "rcx", "rdx", "rsi", "rdi", "memory"
   );
}


/*
 *-----------------------------------------------------------------------------
 *
 * BackdoorHbIn  --
 * BackdoorHbOut --
 *
 *      Send a high-bandwidth basic request to vmware, and return its
 *      reply.
 *
 * Results:
 *      Host-side response returned in bp IN/OUT parameter.
 *
 * Side-effects:
 *      Pokes the high-bandwidth backdoor port.
 *
 *-----------------------------------------------------------------------------
 */

void
BackdoorHbIn(Backdoor_proto_hb *myBp) // IN/OUT
{
   uint64 dummy;

   __asm__ __volatile__(
        "pushq %%rbp"           "\n\t"
#ifdef __APPLE__
        /*
         * Save %rbx on the stack because the Mac OS GCC doesn't want us to
         * clobber it - it erroneously thinks %rbx is the PIC register.
         * (Radar bug 7304232)
         */
        "pushq %%rbx"           "\n\t"
#endif
        "pushq %%rax"           "\n\t"
        "movq 48(%%rax), %%rbp" "\n\t"
        "movq 40(%%rax), %%rdi" "\n\t"
        "movq 32(%%rax), %%rsi" "\n\t"
        "movq 24(%%rax), %%rdx" "\n\t"
        "movq 16(%%rax), %%rcx" "\n\t"
        "movq  8(%%rax), %%rbx" "\n\t"
        "movq   (%%rax), %%rax" "\n\t"
        "cld"                   "\n\t"
        "rep; insb"             "\n\t"
        "xchgq %%rax, (%%rsp)"  "\n\t"
        "movq %%rbp, 48(%%rax)" "\n\t"
        "movq %%rdi, 40(%%rax)" "\n\t"
        "movq %%rsi, 32(%%rax)" "\n\t"
        "movq %%rdx, 24(%%rax)" "\n\t"
        "movq %%rcx, 16(%%rax)" "\n\t"
        "movq %%rbx,  8(%%rax)" "\n\t"
        "popq          (%%rax)" "\n\t"
#ifdef __APPLE__
        "popq %%rbx"            "\n\t"
#endif
        "popq %%rbp"
      : "=a" (dummy)
      : "0" (myBp)
      /*
       * vmware can modify the whole VM state without the compiler knowing
       * it. --hpreg
       */
      :
#ifndef __APPLE__
      /* %rbx is unchanged at the end of the function on Mac OS. */
      "rbx",
#endif
      "rcx", "rdx", "rsi", "rdi", "memory", "cc"
   );
}


void
BackdoorHbOut(Backdoor_proto_hb *myBp) // IN/OUT
{
   uint64 dummy;

   __asm__ __volatile__(
        "pushq %%rbp"           "\n\t"
#ifdef __APPLE__
        /*
         * Save %rbx on the stack because the Mac OS GCC doesn't want us to
         * clobber it - it erroneously thinks %rbx is the PIC register.
         * (Radar bug 7304232)
         */
        "pushq %%rbx"           "\n\t"
#endif
        "pushq %%rax"           "\n\t"
        "movq 48(%%rax), %%rbp" "\n\t"
        "movq 40(%%rax), %%rdi" "\n\t"
        "movq 32(%%rax), %%rsi" "\n\t"
        "movq 24(%%rax), %%rdx" "\n\t"
        "movq 16(%%rax), %%rcx" "\n\t"
        "movq  8(%%rax), %%rbx" "\n\t"
        "movq   (%%rax), %%rax" "\n\t"
        "cld"                   "\n\t"
        "rep; outsb"            "\n\t"
        "xchgq %%rax, (%%rsp)"  "\n\t"
        "movq %%rbp, 48(%%rax)" "\n\t"
        "movq %%rdi, 40(%%rax)" "\n\t"
        "movq %%rsi, 32(%%rax)" "\n\t"
        "movq %%rdx, 24(%%rax)" "\n\t"
        "movq %%rcx, 16(%%rax)" "\n\t"
        "movq %%rbx,  8(%%rax)" "\n\t"
        "popq          (%%rax)" "\n\t"
#ifdef __APPLE__
        "popq %%rbx"            "\n\t"
#endif
        "popq %%rbp"
      : "=a" (dummy)
      : "0" (myBp)
      :
#ifndef __APPLE__
      /* %rbx is unchanged at the end of the function on Mac OS. */
      "rbx",
#endif
      "rcx", "rdx", "rsi", "rdi", "memory", "cc"
   );
}


#ifdef __cplusplus
}
#endif
