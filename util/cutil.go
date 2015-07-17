// +build linux

package util

/*
#cgo LDFLAGS: -lmount -lblkid -luuid -lselinux
#include<blkid/blkid.h>
#include<libmount/libmount.h>
#include<stdlib.h>
*/
import "C"
import "unsafe"

import (
	"errors"
)

func ResolveDevice(spec string) string {
	cSpec := C.CString(spec)
	defer C.free(unsafe.Pointer(cSpec))
	cString := C.blkid_evaluate_spec(cSpec, nil)
	defer C.free(unsafe.Pointer(cString))
	return C.GoString(cString)
}

func GetFsType(device string) (string, error) {
	var ambi *C.int
	cDevice := C.CString(device)
	defer C.free(unsafe.Pointer(cDevice))
	cString := C.mnt_get_fstype(cDevice, ambi, nil)
	defer C.free(unsafe.Pointer(cString))
	if cString != nil {
		return C.GoString(cString), nil
	}
	return "", errors.New("Error while getting fstype")
}

func intToBool(value C.int) bool {
	if value == 0 {
		return false
	}
	return true
}
