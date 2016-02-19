package selinux

// #cgo pkg-config: libselinux
// #include <selinux/selinux.h>
import "C"

func SetFileContext(path string, context string) (int, error) {
	ret, err := C.setfilecon(C.CString(path), C.CString(context))
	return int(ret), err
}
