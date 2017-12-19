/*********************************************************
 * Copyright (C) 2009-2015 VMware, Inc. All rights reserved.
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
 * community_source.h --
 *
 *    Macros for excluding source code from community.
 */

#ifndef _COMMUNITY_SOURCE_H_
#define _COMMUNITY_SOURCE_H_

/* 
 * Convenience macro for COMMUNITY_SOURCE
 */
#undef EXCLUDE_COMMUNITY_SOURCE
#ifdef COMMUNITY_SOURCE
   #define EXCLUDE_COMMUNITY_SOURCE(x) 
#else
   #define EXCLUDE_COMMUNITY_SOURCE(x) x
#endif

#undef COMMUNITY_SOURCE_AMD_SECRET
#if !defined(COMMUNITY_SOURCE) || defined(AMD_SOURCE)
/*
 * It's ok to include AMD_SECRET source code for non-Community Source,
 * or for drops directed at AMD.
 */
   #define COMMUNITY_SOURCE_AMD_SECRET
#endif

#undef COMMUNITY_SOURCE_INTEL_SECRET
#if !defined(COMMUNITY_SOURCE) || defined(INTEL_SOURCE)
/*
 * It's ok to include INTEL_SECRET source code for non-Community Source,
 * or for drops directed at Intel.
 */
   #define COMMUNITY_SOURCE_INTEL_SECRET
#endif

#endif
