package bridge

/*
#cgo CFLAGS: -I../include
#include <stdlib.h>
#include "message.h"
void Warning(const char *fmt, ...) {}
void Debug(const char *fmt, ...) {}
void Panic(const char *fmt, ...) {}
void Log(const char *fmt, ...) {}
*/
import "C"
import "unsafe"

// MessageChannel provides a channel to pass information from/to the hypervisor
type MessageChannel *C.struct_Message_Channel

// MessageOpen creates a new MessageChannel
func MessageOpen(proto uint32) MessageChannel {
	return C.Message_Open(C.uint32(proto))
}

// MessageClose closes a MessageChannel
func MessageClose(c MessageChannel) bool {
	status := C.Message_Close(c)
	return status != 0
}

// MessageSend sends a request through a MessageChannel
func MessageSend(c MessageChannel, request []byte) bool {
	buffer := (*C.uchar)(unsafe.Pointer(&request[0]))
	status := C.Message_Send(c, buffer, (C.size_t)(C.int(len(request))))
	return status != 0
}

// MessageReceive receives a response through a MessageChannel
func MessageReceive(c MessageChannel) ([]byte, bool) {
	var reply *C.uchar
	var replyLen C.size_t
	defer C.free(unsafe.Pointer(reply))

	status := C.Message_Receive(c, &reply, &replyLen)

	res := C.GoBytes(unsafe.Pointer(reply), (C.int)(replyLen))
	return res, status != 0
}
