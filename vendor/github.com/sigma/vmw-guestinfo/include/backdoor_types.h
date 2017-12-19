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
 * backdoor_types.h --
 *
 *    Type definitions for backdoor interaction code.
 */

#ifndef _BACKDOOR_TYPES_H_
#define _BACKDOOR_TYPES_H_

#ifndef VM_I386
#error The backdoor protocol is only supported on x86 architectures.
#endif

/*
 * These #defines are intended for defining register structs as part of
 * existing named unions. If the union should encapsulate the register
 * (and nothing else), use DECLARE_REG_NAMED_STRUCT defined below.
 */

#define DECLARE_REG32_STRUCT \
   struct { \
      uint16 low; \
      uint16 high; \
   } halfs; \
   uint32 word

#define DECLARE_REG64_STRUCT \
   DECLARE_REG32_STRUCT; \
   struct { \
      uint32 low; \
      uint32 high; \
   } words; \
   uint64 quad

#ifndef VM_X86_64
#define DECLARE_REG_STRUCT DECLARE_REG32_STRUCT
#else
#define DECLARE_REG_STRUCT DECLARE_REG64_STRUCT
#endif

#define DECLARE_REG_NAMED_STRUCT(_r) \
   union { DECLARE_REG_STRUCT; } _r

/*
 * Some of the registers are expressed by semantic name, because if they were
 * expressed as register structs declared above, we could only address them
 * by fixed size (half-word, word, quad, etc.) instead of by varying size
 * (size_t, uintptr_t).
 *
 * To be cleaner, these registers are expressed ONLY by semantic name,
 * rather than by a union of the semantic name and a register struct.
 */
typedef union {
   struct {
      DECLARE_REG_NAMED_STRUCT(ax);
      size_t size; /* Register bx. */
      DECLARE_REG_NAMED_STRUCT(cx);
      DECLARE_REG_NAMED_STRUCT(dx);
      DECLARE_REG_NAMED_STRUCT(si);
      DECLARE_REG_NAMED_STRUCT(di);
   } in;
   struct {
      DECLARE_REG_NAMED_STRUCT(ax);
      DECLARE_REG_NAMED_STRUCT(bx);
      DECLARE_REG_NAMED_STRUCT(cx);
      DECLARE_REG_NAMED_STRUCT(dx);
      DECLARE_REG_NAMED_STRUCT(si);
      DECLARE_REG_NAMED_STRUCT(di);
   } out;
} Backdoor_proto;

typedef union {
   struct {
      DECLARE_REG_NAMED_STRUCT(ax);
      DECLARE_REG_NAMED_STRUCT(bx);
      size_t size; /* Register cx. */
      DECLARE_REG_NAMED_STRUCT(dx);
      uintptr_t srcAddr; /* Register si. */
      uintptr_t dstAddr; /* Register di. */
      DECLARE_REG_NAMED_STRUCT(bp);
   } in;
   struct {
      DECLARE_REG_NAMED_STRUCT(ax);
      DECLARE_REG_NAMED_STRUCT(bx);
      DECLARE_REG_NAMED_STRUCT(cx);
      DECLARE_REG_NAMED_STRUCT(dx);
      DECLARE_REG_NAMED_STRUCT(si);
      DECLARE_REG_NAMED_STRUCT(di);
      DECLARE_REG_NAMED_STRUCT(bp);
   } out;
} Backdoor_proto_hb;

MY_ASSERTS(BACKDOOR_STRUCT_SIZES,
           ASSERT_ON_COMPILE(sizeof(Backdoor_proto) == 6 * sizeof(uintptr_t));
           ASSERT_ON_COMPILE(sizeof(Backdoor_proto_hb) == 7 * sizeof(uintptr_t));
)

#undef DECLARE_REG_STRUCT

#endif /* _BACKDOOR_TYPES_H_ */
