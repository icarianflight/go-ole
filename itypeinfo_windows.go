//go:build windows

package ole

import (
	"syscall"
	"unsafe"
)

func (v *ITypeInfo) GetTypeAttr() (tattr *TYPEATTR, err error) {
	hr, _, _ := syscall.SyscallN(
		uintptr(v.VTable().GetTypeAttr),
		uintptr(unsafe.Pointer(v)),
		uintptr(unsafe.Pointer(&tattr)))
	if hr != 0 {
		err = NewError(hr)
	}
	return
}
