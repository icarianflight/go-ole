package ole

import (
	"unicode/utf16"
	"unsafe"
)

// ClassIDFrom retrieves class ID whether given is program ID or application string.
//
// Helper that provides check against both Class ID from Program ID and Class ID from string. It is
// faster, if you know which you are using, to use the individual functions, but this will check
// against available functions for you.
func ClassIDFrom(programID string) (classID *GUID, err error) {
	classID, err = CLSIDFromProgID(programID)
	if err != nil {
		classID, err = CLSIDFromString(programID)
		if err != nil {
			return
		}
	}
	return
}

// BytePtrToString converts byte pointer to a Go string.
func BytePtrToString(p *byte) string {
	a := (*[10000]uint8)(unsafe.Pointer(p))
	i := 0
	for a[i] != 0 {
		i++
	}
	return string(a[:i])
}

// UTF16PtrToString is alias for LpOleStrToString.
//
// Kept for compatibility reasons.
func UTF16PtrToString(p *uint16) string {
	return LpOleStrToString(p)
}

// LpOleStrToString converts COM Unicode to Go string.
func LpOleStrToString(p *uint16) string {
	if p == nil {
		return ""
	}

	length := lpOleStrLen(p)
	a := make([]uint16, length)

	ptr := unsafe.Pointer(p)

	for i := 0; i < int(length); i++ {
		a[i] = *(*uint16)(ptr)
		ptr = unsafe.Pointer(uintptr(ptr) + 2)
	}

	return string(utf16.Decode(a))
}

// BstrToString converts COM binary string to Go string.
func BstrToString(p *uint16) string {
	if p == nil {
		return ""
	}
	length := SysStringLen((*int16)(unsafe.Pointer(p)))
	a := make([]uint16, length)

	ptr := unsafe.Pointer(p)

	for i := 0; i < int(length); i++ {
		a[i] = *(*uint16)(ptr)
		ptr = unsafe.Pointer(uintptr(ptr) + 2)
	}
	return string(utf16.Decode(a))
}

// lpOleStrLen returns the length of Unicode string.
func lpOleStrLen(p *uint16) (length int64) {
	if p == nil {
		return 0
	}

	ptr := unsafe.Pointer(p)

	for i := 0; ; i++ {
		if *(*uint16)(ptr) == 0 {
			length = int64(i)
			break
		}
		ptr = unsafe.Pointer(uintptr(ptr) + 2)
	}
	return
}

// uintptrToPointer reinterprets an address that COM or the Windows API handed
// back as an integer (a syscall return value, a VARIANT.Val slot, an HSTRING
// buffer) as an unsafe.Pointer.
//
// It is intended for memory the Go runtime does not manage (tests deliberately
// use it on Go-allocated values they keep alive themselves): the garbage collector
// neither moves nor frees it, so the unsafe.Pointer rule against converting a
// uintptr back to a Pointer has nothing to protect. Reading the integer's bits
// through a *unsafe.Pointer yields exactly the value a direct
// unsafe.Pointer(p) conversion would, without the pattern go vet rejects. It
// is the same reinterpretation VARIANT.ToString already does on Val.
//
// Do not "simplify" this to unsafe.Add(nil, p): the compiler treats a pointer
// derived from nil as never pointing into the Go heap and drops the GC write
// barrier wherever the result is stored, which the original conversion kept.
func uintptrToPointer(p uintptr) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&p))
}

// convertHresultToError converts syscall to error, if call is unsuccessful.
func convertHresultToError(hr uintptr, r2 uintptr, ignore error) (err error) {
	if hr != 0 {
		err = NewError(hr)
	}
	return
}
