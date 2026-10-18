//go:build windows

package ole

import (
	"syscall"
	"unsafe"
)

func getClassInfo(disp *IProvideClassInfo) (tinfo *ITypeInfo, err error) {
	hr, _, _ := syscall.SyscallN(
		disp.VTable().GetClassInfo,
		uintptr(unsafe.Pointer(disp)),
		uintptr(unsafe.Pointer(&tinfo)))
	if hr != 0 {
		err = NewError(hr)
	}
	return
}
