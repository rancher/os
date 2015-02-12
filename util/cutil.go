package util

/*
#cgo LDFLAGS: -lblkid -luuid
#include<blkid/blkid.h>
#include<stdlib.h>
*/
import "C"
import "unsafe"

func ResolveDevice(spec string) string {
	cString := C.blkid_evaluate_spec(C.CString(spec), nil)
	defer C.free(unsafe.Pointer(cString)) 
	return C.GoString(cString)
}

