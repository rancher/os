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
	"bytes"
	"encoding/binary"
	"errors"
	"unsafe"

	"github.com/vmware/vmw-guestinfo/bdoor"
)

const (
	messageTypeOpen = iota
	messageTypeSendSize
	messageTypeSendPayload
	messageTypeReceiveSize
	messageTypeReceivePayload
	messageTypeReceiveStatus
	messageTypeClose

	messageStatusFail       = uint16(0x0000)
	messageStatusSuccess    = uint16(0x0001)
	messageStatusDoRecieve  = uint16(0x0002)
	messageStatusCheckPoint = uint16(0x0010)
	messageStatusHighBW     = uint16(0x0080)
)

var (
	// ErrChannelOpen represents a failure to open a channel
	ErrChannelOpen = errors.New("could not open channel")
	// ErrChannelClose represents a failure to close a channel
	ErrChannelClose = errors.New("could not close channel")
	// ErrRpciSend represents a failure to send a message
	ErrRpciSend = errors.New("unable to send RPCI command")
	// ErrRpciReceive represents a failure to receive a message
	ErrRpciReceive = errors.New("unable to receive RPCI command result")
)

type Channel struct {
	id uint16

	forceLowBW bool
	buf        []byte

	cookie bdoor.UInt64
}

// NewChannel opens a new Channel
func NewChannel(proto uint32) (*Channel, error) {
	flags := bdoor.CommandFlagCookie

retry:
	bp := &bdoor.BackdoorProto{}

	bp.BX.Low.SetWord(proto | flags)
	bp.CX.Low.High = messageTypeOpen
	bp.CX.Low.Low = bdoor.CommandMessage

	out := bp.InOut()
	if (out.CX.Low.High & messageStatusSuccess) == 0 {
		if flags != 0 {
			flags = 0
			goto retry
		}

		Errorf("Message: Unable to open communication channel")
		return nil, ErrChannelOpen
	}

	ch := &Channel{}
	ch.id = out.DX.Low.High
	ch.cookie.High.SetWord(out.SI.Low.Word())
	ch.cookie.Low.SetWord(out.DI.Low.Word())

	Debugf("Opened channel %d", ch.id)
	return ch, nil
}

func (c *Channel) Close() error {
	bp := &bdoor.BackdoorProto{}

	bp.CX.Low.High = messageTypeClose
	bp.CX.Low.Low = bdoor.CommandMessage

	bp.DX.Low.High = c.id
	bp.SI.Low.SetWord(c.cookie.High.Word())
	bp.DI.Low.SetWord(c.cookie.Low.Word())

	out := bp.InOut()
	if (out.CX.Low.High & messageStatusSuccess) == 0 {
		Errorf("Message: Unable to close communication channel %d", c.id)
		return ErrChannelClose
	}

	Debugf("Closed channel %d", c.id)
	return nil
}

func (c *Channel) Send(buf []byte) error {
retry:
	bp := &bdoor.BackdoorProto{}
	bp.CX.Low.High = messageTypeSendSize
	bp.CX.Low.Low = bdoor.CommandMessage

	bp.DX.Low.High = c.id
	bp.SI.Low.SetWord(c.cookie.High.Word())
	bp.DI.Low.SetWord(c.cookie.Low.Word())

	bp.BX.Low.SetWord(uint32(len(buf)))

	// send the size
	out := bp.InOut()
	if (out.CX.Low.High & messageStatusSuccess) == 0 {
		Errorf("Message: Unable to send a message over the communication channel %d", c.id)
		return ErrRpciSend
	}

	// size of buf 0 is fine, just return
	if len(buf) == 0 {
		return nil
	}

	if !c.forceLowBW && (out.CX.Low.High&messageStatusHighBW) == messageStatusHighBW {
		hbbp := &bdoor.BackdoorProto{}

		hbbp.BX.Low.Low = bdoor.CommandHighBWMessage
		hbbp.BX.Low.High = messageStatusSuccess
		hbbp.DX.Low.High = c.id
		hbbp.BP.Low.SetWord(c.cookie.High.Word())
		hbbp.DI.Low.SetWord(c.cookie.Low.Word())
		hbbp.CX.Low.SetWord(uint32(len(buf)))
		hbbp.SI.SetQuad(uint64(uintptr(unsafe.Pointer(&buf[0]))))

		out := hbbp.HighBandwidthOut()
		if (out.BX.Low.High & messageStatusSuccess) == 0 {
			if (out.BX.Low.High & messageStatusCheckPoint) != 0 {
				Debugf("A checkpoint occurred. Retrying the operation")
				goto retry
			}

			Errorf("Message: Unable to send a message over the communication channel %d", c.id)
			return ErrRpciSend
		}
	} else {
		bp.CX.Low.High = messageTypeSendPayload

		bbuf := bytes.NewBuffer(buf)
		for {
			// read 4 bytes at a time
			words := bbuf.Next(4)
			if len(words) == 0 {
				break
			}

			Debugf("sending %q over %d", string(words), c.id)
			switch len(words) {
			case 3:
				bp.BX.Low.SetWord(binary.LittleEndian.Uint32([]byte{0x0, words[2], words[1], words[0]}))
			case 2:
				bp.BX.Low.SetWord(uint32(binary.LittleEndian.Uint16(words)))
			case 1:
				bp.BX.Low.SetWord(uint32(words[0]))
			default:
				bp.BX.Low.SetWord(binary.LittleEndian.Uint32(words))
			}

			out = bp.InOut()
			if (out.CX.Low.High & messageStatusSuccess) == 0 {
				Errorf("Message: Unable to send a message over the communication channel %d", c.id)
				return ErrRpciSend
			}
		}
	}

	return nil
}

func (c *Channel) Receive() ([]byte, error) {
retry:
	var err error
	bp := &bdoor.BackdoorProto{}
	bp.CX.Low.High = messageTypeReceiveSize
	bp.CX.Low.Low = bdoor.CommandMessage

	bp.DX.Low.High = c.id
	bp.SI.Low.SetWord(c.cookie.High.Word())
	bp.DI.Low.SetWord(c.cookie.Low.Word())

	out := bp.InOut()
	if (out.CX.Low.High & messageStatusSuccess) == 0 {
		Errorf("Message: Unable to poll for messages over the communication channel %d", c.id)
		return nil, ErrRpciReceive
	}

	if (out.CX.Low.High & messageStatusDoRecieve) == 0 {
		Debugf("No message to retrieve")
		return nil, nil
	}

	// Receive the size.
	if out.DX.Low.High != messageTypeSendSize {
		Errorf("Message: Protocol error. Expected a MESSAGE_TYPE_SENDSIZE request from vmware")
		return nil, ErrRpciReceive
	}

	size := out.BX.Quad()

	var buf []byte

	if size != 0 {
		if !c.forceLowBW && (out.CX.Low.High&messageStatusHighBW == messageStatusHighBW) {
			buf = make([]byte, size)

			hbbp := &bdoor.BackdoorProto{}

			hbbp.BX.Low.Low = bdoor.CommandHighBWMessage
			hbbp.BX.Low.High = messageStatusSuccess
			hbbp.DX.Low.High = c.id
			hbbp.SI.Low.SetWord(c.cookie.High.Word())
			hbbp.BP.Low.SetWord(c.cookie.Low.Word())
			hbbp.CX.Low.SetWord(uint32(len(buf)))
			hbbp.DI.SetQuad(uint64(uintptr(unsafe.Pointer(&buf[0]))))

			out := hbbp.HighBandwidthIn()
			if (out.BX.Low.High & messageStatusSuccess) == 0 {
				Errorf("Message: Unable to send a message over the communication channel %d", c.id)
				c.reply(messageTypeReceivePayload, messageStatusFail)
				return nil, ErrRpciReceive
			}
		} else {
			b := bytes.NewBuffer(make([]byte, 0, size))

			for {
				if size == 0 {
					break
				}

				bp.CX.Low.High = messageTypeReceivePayload
				bp.BX.Low.Low = messageStatusSuccess

				out = bp.InOut()
				if (out.CX.Low.High & messageStatusSuccess) == 0 {
					if (out.CX.Low.High & messageStatusCheckPoint) != 0 {
						Debugf("A checkpoint occurred. Retrying the operation")
						goto retry
					}

					Errorf("Message: Unable to receive a message over the communication channel %d", c.id)
					c.reply(messageTypeReceivePayload, messageStatusFail)
					return nil, ErrRpciReceive
				}

				if out.DX.Low.High != messageTypeSendPayload {
					Errorf("Message: Protocol error. Expected a MESSAGE_TYPE_SENDPAYLOAD from vmware")
					c.reply(messageTypeReceivePayload, messageStatusFail)
					return nil, ErrRpciReceive
				}

				Debugf("Received %#v", out.BX.Low.Word())

				switch size {
				case 1:
					err = binary.Write(b, binary.LittleEndian, uint8(out.BX.Low.Low))
					size = size - 1

				case 2:
					err = binary.Write(b, binary.LittleEndian, uint16(out.BX.Low.Low))
					size = size - 2

				case 3:
					err = binary.Write(b, binary.LittleEndian, uint16(out.BX.Low.Low))
					if err != nil {
						c.reply(messageTypeReceivePayload, messageStatusFail)
						return nil, ErrRpciReceive
					}
					err = binary.Write(b, binary.LittleEndian, uint8(out.BX.Low.High))
					size = size - 3

				default:
					err = binary.Write(b, binary.LittleEndian, out.BX.Low.Word())
					size = size - 4
				}

				if err != nil {
					Errorf(err.Error())
					c.reply(messageTypeReceivePayload, messageStatusFail)
					return nil, ErrRpciReceive
				}
			}

			buf = b.Bytes()
		}
	}

	c.reply(messageTypeReceiveStatus, messageStatusSuccess)

	return buf, nil
}

func (c *Channel) reply(messageType, messageStatus uint16) {
	bp := &bdoor.BackdoorProto{}

	bp.BX.Low.Low = messageStatus
	bp.CX.Low.High = messageType
	bp.CX.Low.Low = bdoor.CommandMessage
	bp.DX.Low.High = c.id
	bp.SI.Low.SetWord(c.cookie.High.Word())
	bp.DI.Low.SetWord(c.cookie.Low.Word())

	out := bp.InOut()

	/* OUT: Status */
	if (out.CX.Low.High & messageStatusSuccess) == 0 {
		if messageStatus == messageStatusSuccess {
			Errorf("reply Message: Unable to send a message over the communication channel %d", c.id)
		} else {
			Errorf("reply Message: Unable to signal an error of reception over the communication channel %d", c.id)
		}
	}
}
