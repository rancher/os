//go:build linux
// +build linux

package util

/*
#cgo LDFLAGS: -lblkid -luuid
#include<blkid/blkid.h>
#include<stdlib.h>
*/
import "C"
import (
	"errors"
	"unsafe"
)

// ResolveDevice this isn't reliable - blkid -L LABEL works more often :(
func ResolveDevice(spec string) string {
	cSpec := C.CString(spec)
	defer C.free(unsafe.Pointer(cSpec))
	cString := C.blkid_evaluate_spec(cSpec, nil)
	defer C.free(unsafe.Pointer(cString))
	return C.GoString(cString)
}

func GetFsType(device string) (string, error) {
	cDevice := C.CString(device)
	defer C.free(unsafe.Pointer(cDevice))
	cTag := C.CString("TYPE")
	defer C.free(unsafe.Pointer(cTag))
	cString := C.blkid_get_tag_value(nil, cTag, cDevice)
	defer C.free(unsafe.Pointer(cString))
	if cString != nil {
		return C.GoString(cString), nil
	}
	return "", errors.New("Error while getting fstype")
}
