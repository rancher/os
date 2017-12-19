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
 * guest_msg_def.h --
 *
 *    Second layer of the internal communication channel between guest
 *    applications and vmware
 *
 */

#ifndef _GUEST_MSG_DEF_H_
#define _GUEST_MSG_DEF_H_

/* Basic request types */
typedef enum {
   MESSAGE_TYPE_OPEN,
   MESSAGE_TYPE_SENDSIZE,
   MESSAGE_TYPE_SENDPAYLOAD,
   MESSAGE_TYPE_RECVSIZE,
   MESSAGE_TYPE_RECVPAYLOAD,
   MESSAGE_TYPE_RECVSTATUS,
   MESSAGE_TYPE_CLOSE,
} MessageType;


/* Reply statuses */
/*  The basic request succeeded */
#define MESSAGE_STATUS_SUCCESS  0x0001
/*  vmware has a message available for its party */
#define MESSAGE_STATUS_DORECV   0x0002
/*  The channel has been closed */
#define MESSAGE_STATUS_CLOSED   0x0004
/*  vmware removed the message before the party fetched it */
#define MESSAGE_STATUS_UNSENT   0x0008
/*  A checkpoint occurred */
#define MESSAGE_STATUS_CPT      0x0010
/*  An underlying device is powering off */
#define MESSAGE_STATUS_POWEROFF 0x0020
/*  vmware has detected a timeout on the channel */
#define MESSAGE_STATUS_TIMEOUT  0x0040
/*  vmware supports high-bandwidth for sending and receiving the payload */
#define MESSAGE_STATUS_HB       0x0080

/*
 * This mask defines the status bits that the guest is allowed to set;
 * we use this to mask out all other bits when receiving the status
 * from the guest. Otherwise, the guest can manipulate VMX state by
 * setting status bits that are only supposed to be changed by the
 * VMX. See bug 45385.
 */
#define MESSAGE_STATUS_GUEST_MASK    MESSAGE_STATUS_SUCCESS

/*
 * Max number of channels.
 * Unfortunately this has to be public because the monitor part
 * of the backdoor needs it for its trivial-case optimization. [greg]
 */
#define GUESTMSG_MAX_CHANNEL 8

/* Flags to open a channel. --hpreg */
#define GUESTMSG_FLAG_COOKIE 0x80000000
#define GUESTMSG_FLAG_ALL GUESTMSG_FLAG_COOKIE

/*
 * Maximum size of incoming message. This is to prevent denial of host service
 * attacks from guest applications.
 */
#define GUESTMSG_MAX_IN_SIZE (64 * 1024)

#endif /* _GUEST_MSG_DEF_H_ */
