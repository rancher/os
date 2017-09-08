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
 * backdoorGcc32.c --
 *
 *      Implements the real work for guest-side backdoor for GCC, 32-bit
 *      target (supports inline ASM, GAS syntax). The asm sections are marked
 *      volatile since vmware can change the registers content without the
 *      compiler knowing it.
 *
 *      XXX
 *      I tried to write this more cleanly, but:
 *        - There is no way to specify an "ebp" constraint
 *        - "ebp" is ignored when specified as cloberred register
 *        - gas barfs when there is more than 10 operands
 *        - gas 2.7.2.3, depending on the order of the operands, can
 *          mis-assemble without any warning
 *      --hpreg
 *
 *      Note that the problems with gas noted above might longer be relevant
 *      now that we've upgraded most of our compiler versions.
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
   uint32 dummy;

   __asm__ __volatile__(
#ifdef __PIC__
        "pushl %%ebx"           "\n\t"
#endif
        "pushl %%eax"           "\n\t"
        "movl 20(%%eax), %%edi" "\n\t"
        "movl 16(%%eax), %%esi" "\n\t"
        "movl 12(%%eax), %%edx" "\n\t"
        "movl  8(%%eax), %%ecx" "\n\t"
        "movl  4(%%eax), %%ebx" "\n\t"
        "movl   (%%eax), %%eax" "\n\t"
        "inl %%dx, %%eax"       "\n\t"
        "xchgl %%eax, (%%esp)"  "\n\t"
        "movl %%edi, 20(%%eax)" "\n\t"
        "movl %%esi, 16(%%eax)" "\n\t"
        "movl %%edx, 12(%%eax)" "\n\t"
        "movl %%ecx,  8(%%eax)" "\n\t"
        "movl %%ebx,  4(%%eax)" "\n\t"
        "popl          (%%eax)" "\n\t"
#ifdef __PIC__
        "popl %%ebx"            "\n\t"
#endif
      : "=a" (dummy)
      : "0" (myBp)
      /*
       * vmware can modify the whole VM state without the compiler knowing
       * it. So far it does not modify EFLAGS. --hpreg
       */
      :
#ifndef __PIC__
        "ebx",
#endif
        "ecx", "edx", "esi", "edi", "memory"
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
   uint32 dummy;

   __asm__ __volatile__(
#ifdef __PIC__
        "pushl %%ebx"           "\n\t"
#endif
        "pushl %%ebp"           "\n\t"

        "pushl %%eax"           "\n\t"
        "movl 24(%%eax), %%ebp" "\n\t"
        "movl 20(%%eax), %%edi" "\n\t"
        "movl 16(%%eax), %%esi" "\n\t"
        "movl 12(%%eax), %%edx" "\n\t"
        "movl  8(%%eax), %%ecx" "\n\t"
        "movl  4(%%eax), %%ebx" "\n\t"
        "movl   (%%eax), %%eax" "\n\t"
        "cld"                   "\n\t"
        "rep; insb"             "\n\t"
        "xchgl %%eax, (%%esp)"  "\n\t"
        "movl %%ebp, 24(%%eax)" "\n\t"
        "movl %%edi, 20(%%eax)" "\n\t"
        "movl %%esi, 16(%%eax)" "\n\t"
        "movl %%edx, 12(%%eax)" "\n\t"
        "movl %%ecx,  8(%%eax)" "\n\t"
        "movl %%ebx,  4(%%eax)" "\n\t"
        "popl          (%%eax)" "\n\t"

        "popl %%ebp"            "\n\t"
#ifdef __PIC__
        "popl %%ebx"            "\n\t"
#endif
      : "=a" (dummy)
      : "0" (myBp)
      /*
       * vmware can modify the whole VM state without the compiler knowing
       * it. --hpreg
       */
      : 
#ifndef __PIC__
        "ebx", 
#endif
        "ecx", "edx", "esi", "edi", "memory", "cc"
   );
}


void
BackdoorHbOut(Backdoor_proto_hb *myBp) // IN/OUT
{
   uint32 dummy;

   __asm__ __volatile__(
#ifdef __PIC__
        "pushl %%ebx"           "\n\t"
#endif
        "pushl %%ebp"           "\n\t"

        "pushl %%eax"           "\n\t"
        "movl 24(%%eax), %%ebp" "\n\t"
        "movl 20(%%eax), %%edi" "\n\t"
        "movl 16(%%eax), %%esi" "\n\t"
        "movl 12(%%eax), %%edx" "\n\t"
        "movl  8(%%eax), %%ecx" "\n\t"
        "movl  4(%%eax), %%ebx" "\n\t"
        "movl   (%%eax), %%eax" "\n\t"
        "cld"                   "\n\t"
        "rep; outsb"            "\n\t"
        "xchgl %%eax, (%%esp)"  "\n\t"
        "movl %%ebp, 24(%%eax)" "\n\t"
        "movl %%edi, 20(%%eax)" "\n\t"
        "movl %%esi, 16(%%eax)" "\n\t"
        "movl %%edx, 12(%%eax)" "\n\t"
        "movl %%ecx,  8(%%eax)" "\n\t"
        "movl %%ebx,  4(%%eax)" "\n\t"
        "popl          (%%eax)" "\n\t"

        "popl %%ebp"            "\n\t"
#ifdef __PIC__
        "popl %%ebx"            "\n\t"
#endif
      : "=a" (dummy)
      : "0" (myBp)
      :
#ifndef __PIC__
        "ebx",
#endif
        "ecx", "edx", "esi", "edi", "memory", "cc"
   );
}

#ifdef __cplusplus
}
#endif

