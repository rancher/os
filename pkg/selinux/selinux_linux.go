package selinux

// #cgo pkg-config: libselinux libsepol
// #include <selinux/selinux.h>
// #include <stdlib.h>
import "C"

import (
	"unsafe"
)

func InitializeSelinux() (int, error) {
	enforce := C.int(0)
	ret, err := C.selinux_init_load_policy(&enforce)
	return int(ret), err
}

func SetFileContext(path string, context string) (int, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	cContext := C.CString(context)
	defer C.free(unsafe.Pointer(cContext))

	ret, err := C.setfilecon(cPath, cContext)
	return int(ret), err
}
