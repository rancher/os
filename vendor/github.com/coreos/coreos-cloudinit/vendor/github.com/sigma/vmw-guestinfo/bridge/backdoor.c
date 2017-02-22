/*********************************************************
 * Copyright (C) 1999-2015 VMware, Inc. All rights reserved.
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
 * backdoor.c --
 *
 *    First layer of the internal communication channel between guest
 *    applications and vmware
 *
 *    This is the backdoor. By using special ports of the virtual I/O space,
 *    and the virtual CPU registers, a guest application can send a
 *    synchroneous basic request to vmware, and vmware can reply to it.
 */

#ifdef __cplusplus
extern "C" {
#endif

#include "backdoor_def.h"
#include "backdoor.h"
#include "backdoorInt.h"

#if defined(BACKDOOR_DEBUG) && defined(USERLEVEL)
#if defined(__KERNEL__) || defined(_KERNEL)
#else
#   include "debug.h"
#endif
#   include <stdio.h>
#   define BACKDOOR_LOG(args) Debug args
#   define BACKDOOR_LOG_PROTO_STRUCT(x) BackdoorPrintProtoStruct((x))
#   define BACKDOOR_LOG_HB_PROTO_STRUCT(x) BackdoorPrintHbProtoStruct((x))


/*
 *----------------------------------------------------------------------------
 *
 * BackdoorPrintProtoStruct --
 * BackdoorPrintHbProtoStruct --
 *
 *      Print the contents of the specified backdoor protocol structure via
 *      printf.
 *
 * Results:
 *      None.
 *
 * Side effects:
 *      Output to stdout.
 *
 *----------------------------------------------------------------------------
 */

void
BackdoorPrintProtoStruct(Backdoor_proto *myBp)
{
   Debug("magic 0x%08x, command %d, size %"FMTSZ"u, port %d\n",
         myBp->in.ax.word, myBp->in.cx.halfs.low,
         myBp->in.size, myBp->in.dx.halfs.low);

#ifndef VM_X86_64
   Debug("ax %#x, "
         "bx %#x, "
         "cx %#x, "
         "dx %#x, "
         "si %#x, "
         "di %#x\n",
         myBp->out.ax.word,
         myBp->out.bx.word,
         myBp->out.cx.word,
         myBp->out.dx.word,
         myBp->out.si.word,
         myBp->out.di.word);
#else
   Debug("ax %#"FMT64"x, "
         "bx %#"FMT64"x, "
         "cx %#"FMT64"x, "
         "dx %#"FMT64"x, "
         "si %#"FMT64"x, "
         "di %#"FMT64"x\n",
         myBp->out.ax.quad,
         myBp->out.bx.quad,
         myBp->out.cx.quad,
         myBp->out.dx.quad,
         myBp->out.si.quad,
         myBp->out.di.quad);
#endif
}


void
BackdoorPrintHbProtoStruct(Backdoor_proto_hb *myBp)
{
   Debug("magic 0x%08x, command %d, size %"FMTSZ"u, port %d, "
         "srcAddr %"FMTSZ"u, dstAddr %"FMTSZ"u\n",
         myBp->in.ax.word, myBp->in.bx.halfs.low, myBp->in.size,
         myBp->in.dx.halfs.low, myBp->in.srcAddr, myBp->in.dstAddr);

#ifndef VM_X86_64
   Debug("ax %#x, "
         "bx %#x, "
         "cx %#x, "
         "dx %#x, "
         "si %#x, "
         "di %#x, "
         "bp %#x\n",
         myBp->out.ax.word,
         myBp->out.bx.word,
         myBp->out.cx.word,
         myBp->out.dx.word,
         myBp->out.si.word,
         myBp->out.di.word,
         myBp->out.bp.word);
#else
   Debug("ax %#"FMT64"x, "
         "bx %#"FMT64"x, "
         "cx %#"FMT64"x, "
         "dx %#"FMT64"x, "
         "si %#"FMT64"x, "
         "di %#"FMT64"x, "
         "bp %#"FMT64"x\n",
         myBp->out.ax.quad,
         myBp->out.bx.quad,
         myBp->out.cx.quad,
         myBp->out.dx.quad,
         myBp->out.si.quad,
         myBp->out.di.quad,
         myBp->out.bp.quad);
#endif
}

#else
#   define BACKDOOR_LOG(args)
#   define BACKDOOR_LOG_PROTO_STRUCT(x)
#   define BACKDOOR_LOG_HB_PROTO_STRUCT(x)
#endif


/*
 *-----------------------------------------------------------------------------
 *
 * Backdoor --
 *
 *      Send a low-bandwidth basic request (16 bytes) to vmware, and return its
 *      reply (24 bytes).
 *
 * Result:
 *      None
 *
 * Side-effects:
 *      None
 *
 *-----------------------------------------------------------------------------
 */

void
Backdoor(Backdoor_proto *myBp) // IN/OUT
{
   ASSERT(myBp);

   myBp->in.ax.word = BDOOR_MAGIC;
   myBp->in.dx.halfs.low = BDOOR_PORT;

   BACKDOOR_LOG(("Backdoor: before "));
   BACKDOOR_LOG_PROTO_STRUCT(myBp);

   Backdoor_InOut(myBp);

   BACKDOOR_LOG(("Backdoor: after "));
   BACKDOOR_LOG_PROTO_STRUCT(myBp);
}


/*
 *-----------------------------------------------------------------------------
 *
 * Backdoor_HbOut --
 *
 *      Send a high-bandwidth basic request to vmware, and return its
 *      reply.
 *
 * Result:
 *      The host-side response is returned via the IN/OUT parameter.
 *
 * Side-effects:
 *      Pokes the high-bandwidth backdoor.
 *
 *-----------------------------------------------------------------------------
 */

void
Backdoor_HbOut(Backdoor_proto_hb *myBp) // IN/OUT
{
   ASSERT(myBp);

   myBp->in.ax.word = BDOOR_MAGIC;
   myBp->in.dx.halfs.low = BDOORHB_PORT;

   BACKDOOR_LOG(("Backdoor_HbOut: before "));
   BACKDOOR_LOG_HB_PROTO_STRUCT(myBp);

   BackdoorHbOut(myBp);

   BACKDOOR_LOG(("Backdoor_HbOut: after "));
   BACKDOOR_LOG_HB_PROTO_STRUCT(myBp);
}


/*
 *-----------------------------------------------------------------------------
 *
 * Backdoor_HbIn --
 *
 *      Send a basic request to vmware, and return its high-bandwidth
 *      reply
 *
 * Result:
 *      Host-side response returned via the IN/OUT parameter.
 *
 * Side-effects:
 *      Pokes the high-bandwidth backdoor.
 *
 *-----------------------------------------------------------------------------
 */

void
Backdoor_HbIn(Backdoor_proto_hb *myBp) // IN/OUT
{
   ASSERT(myBp);

   myBp->in.ax.word = BDOOR_MAGIC;
   myBp->in.dx.halfs.low = BDOORHB_PORT;

   BACKDOOR_LOG(("Backdoor_HbIn: before "));
   BACKDOOR_LOG_HB_PROTO_STRUCT(myBp);

   BackdoorHbIn(myBp);

   BACKDOOR_LOG(("Backdoor_HbIn: after "));
   BACKDOOR_LOG_HB_PROTO_STRUCT(myBp);
}

#ifdef __cplusplus
}
#endif
