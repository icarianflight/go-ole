//go:build windows

package ole

import (
	"reflect"
	"syscall"
	"unsafe"
)

func reflectQueryInterface(self any, method uintptr, interfaceID *GUID, obj any) (err error) {
	selfValue := reflect.ValueOf(self).Elem()
	objValue := reflect.ValueOf(obj).Elem()

	hr, _, _ := syscall.SyscallN(
		method,
		selfValue.UnsafeAddr(),
		uintptr(unsafe.Pointer(interfaceID)),
		objValue.Addr().Pointer())
	if hr != 0 {
		err = NewError(hr)
	}
	return
}

func queryInterface(unk *IUnknown, iid *GUID) (disp *IDispatch, err error) {
	hr, _, _ := syscall.SyscallN(
		unk.VTable().QueryInterface,
		uintptr(unsafe.Pointer(unk)),
		uintptr(unsafe.Pointer(iid)),
		uintptr(unsafe.Pointer(&disp)))
	if hr != 0 {
		err = NewError(hr)
	}
	return
}

func addRef(unk *IUnknown) int32 {
	ret, _, _ := syscall.SyscallN(
		unk.VTable().AddRef,
		uintptr(unsafe.Pointer(unk)))
	return int32(ret)
}

func release(unk *IUnknown) int32 {
	ret, _, _ := syscall.SyscallN(
		unk.VTable().Release,
		uintptr(unsafe.Pointer(unk)))
	return int32(ret)
}
