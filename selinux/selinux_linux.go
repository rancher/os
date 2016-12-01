package selinux

// #cgo pkg-config: libselinux libsepol
// #include <selinux/selinux.h>
import "C"

func InitializeSelinux() (int, error) {
	enforce := C.int(0)
	ret, err := C.selinux_init_load_policy(&enforce)
	return int(ret), err
}

func SetFileContext(path string, context string) (int, error) {
	ret, err := C.setfilecon(C.CString(path), C.CString(context))
	return int(ret), err
}
