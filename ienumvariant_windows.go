//go:build windows

package ole

import (
	"syscall"
	"unsafe"
)

func (enum *IEnumVARIANT) Clone() (cloned *IEnumVARIANT, err error) {
	hr, _, _ := syscall.SyscallN(
		enum.VTable().Clone,
		uintptr(unsafe.Pointer(enum)),
		uintptr(unsafe.Pointer(&cloned)))
	if hr != 0 {
		err = NewError(hr)
	}
	return
}

func (enum *IEnumVARIANT) Reset() (err error) {
	hr, _, _ := syscall.SyscallN(
		enum.VTable().Reset,
		uintptr(unsafe.Pointer(enum)))
	if hr != 0 {
		err = NewError(hr)
	}
	return
}

func (enum *IEnumVARIANT) Skip(celt uint) (err error) {
	hr, _, _ := syscall.SyscallN(
		enum.VTable().Skip,
		uintptr(unsafe.Pointer(enum)),
		uintptr(celt))
	if hr != 0 {
		err = NewError(hr)
	}
	return
}

func (enum *IEnumVARIANT) Next(celt uint) (array VARIANT, length uint, err error) {
	hr, _, _ := syscall.SyscallN(
		enum.VTable().Next,
		uintptr(unsafe.Pointer(enum)),
		uintptr(celt),
		uintptr(unsafe.Pointer(&array)),
		uintptr(unsafe.Pointer(&length)))
	if hr != 0 {
		err = NewError(hr)
	}
	return
}
