package message

import (
	"errors"

	"github.com/vmware/vmw-guestinfo/bridge"
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

// Channel is a communication channel between hypervisor and virtual machine
type Channel struct {
	privChan bridge.MessageChannel
}

// NewChannel opens a new channel
func NewChannel(proto uint32) (*Channel, error) {
	if channel := bridge.MessageOpen(proto); channel != nil {
		return &Channel{
			privChan: channel,
		}, nil
	}
	return nil, ErrChannelOpen
}

// Close the channel
func (c *Channel) Close() error {
	if status := bridge.MessageClose(bridge.MessageChannel(c.privChan)); status {
		return nil
	}
	return ErrChannelClose
}

// Send a request
func (c *Channel) Send(request []byte) error {
	if status := bridge.MessageSend(bridge.MessageChannel(c.privChan), request); status {
		return nil
	}
	return ErrRpciSend
}

// Receive a response
func (c *Channel) Receive() ([]byte, error) {
	if res, status := bridge.MessageReceive(bridge.MessageChannel(c.privChan)); status {
		return res, nil
	}
	return nil, ErrRpciReceive
}
