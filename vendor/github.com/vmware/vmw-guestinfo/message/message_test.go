// Copyright 2016 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package message

import (
	"os"
	"testing"

	"github.com/vmware/vmw-guestinfo/util"
	"github.com/vmware/vmw-guestinfo/vmcheck"
)

const rpciProtocolNum uint32 = 0x49435052
const tcloProtocol uint32 = 0x4f4c4354

func TestOpenClose(t *testing.T) {
	l := DefaultLogger.(*logger)
	l.DebugLevel = true

	isVm, err := vmcheck.IsVirtualWorld()
	if err != nil || !isVm {
		t.Skip("Not in a virtual world")
		return
	}

	ch, err := NewChannel(rpciProtocolNum)
	if !util.AssertNotNil(t, ch) || !util.AssertNoError(t, err) {
		return
	}

	// check low bandwidth
	ch.forceLowBW = true
	err = ch.Send([]byte("info-get guestinfo.doesnotexistdoesnotexit"))
	if !util.AssertNoError(t, err) {
		return
	}

	b, err := ch.Receive()
	if !util.AssertNoError(t, err) || !util.AssertNotNil(t, b) {
		return
	}

	if !util.AssertEqual(t, "0 No value found", string(b)) {
		return
	}

	if !util.AssertNoError(t, ch.Close()) {
		return
	}

	// check high bandwidth
	ch, err = NewChannel(rpciProtocolNum)
	if !util.AssertNotNil(t, ch) || !util.AssertNoError(t, err) {
		return
	}

	err = ch.Send([]byte("info-get guestinfo.doesnotexistdoesnotexit"))
	if !util.AssertNoError(t, err) {
		return
	}

	b, err = ch.Receive()
	if !util.AssertNoError(t, err) || !util.AssertNotNil(t, b) {
		return
	}

	if !util.AssertEqual(t, "0 No value found", string(b)) {
		return
	}

	if !util.AssertNoError(t, ch.Close()) {
		return
	}
}

// Test we can reply to the rpcin
func TestReset(t *testing.T) {
	l := DefaultLogger.(*logger)
	l.DebugLevel = true

	isVm, err := vmcheck.IsVirtualWorld()
	if err != nil || !isVm {
		t.Skip("Not in a virtual world")
		return
	}

	if os.Getenv("TEST_TOOLBOX") == "" {
		t.Skip("Skipping toolbox test")
		return
	}

	ch, err := NewChannel(tcloProtocol)
	if !util.AssertNotNil(t, ch) || !util.AssertNoError(t, err) {
		return
	}
	defer ch.Close()

	var buf []byte

	for {
		_ = ch.Send(buf)
		request, _ := ch.Receive()

		if len(request) == 0 {
			continue
		}

		if string(request) == "reset" {
			break
		}
	}

	reply := "OK ATR toolbox"
	err = ch.Send([]byte(reply))
	if !util.AssertNoError(t, err) {
		return
	}
}
