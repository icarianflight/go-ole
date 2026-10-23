//go:build windows

package oleutil

import (
	"syscall"
	"testing"
	"unsafe"

	ole "github.com/icarianflight/go-ole"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A hand-built IDispatch exposing _NewEnum, and a hand-built IEnumVARIANT
// yielding the integers 1..3. Both live entirely in this process.

const fakeEnumItems = 3

type fakeCOMObject struct {
	vtbl unsafe.Pointer
	next int32
}

var (
	fakeEnumObj    *fakeCOMObject
	fakeEnumVtbl   *ole.IEnumVARIANTVtbl
	fakeDispObj    *fakeCOMObject
	fakeDispVtbl   *ole.IDispatchVtbl
	fakeEnumShared uintptr
)

// uintptrToPointer reinterprets an address handed back by COM as a pointer,
// the same way the ole package does. The fake objects are kept alive by
// package-level variables, so the Go runtime never moves or frees them.
func uintptrToPointer(p uintptr) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&p))
}

func fakeQueryInterface(this, iid, ppv uintptr) uintptr {
	*(*uintptr)(uintptrToPointer(ppv)) = this
	return 0
}

func fakeRefCount(this uintptr) uintptr { return 1 }

func fakeEnumNext(this, celt, rgVar, pFetched uintptr) uintptr {
	obj := (*fakeCOMObject)(uintptrToPointer(this))
	if obj.next >= fakeEnumItems {
		*(*uint32)(uintptrToPointer(pFetched)) = 0
		return 1 // S_FALSE
	}
	obj.next++
	v := (*ole.VARIANT)(uintptrToPointer(rgVar))
	v.VT = ole.VT_I4
	v.Val = int64(obj.next)
	*(*uint32)(uintptrToPointer(pFetched)) = 1
	return 0
}

func fakeGetIDsOfNames(this, riid, names, cNames, lcid, dispids uintptr) uintptr {
	*(*int32)(uintptrToPointer(dispids)) = 1
	return 0
}

func fakeInvokeNewEnum(this, dispid, riid, lcid, flags, params, result, excep, argErr uintptr) uintptr {
	v := (*ole.VARIANT)(uintptrToPointer(result))
	v.VT = ole.VT_UNKNOWN
	v.Val = int64(fakeEnumShared)
	return 0
}

func init() {
	qi := syscall.NewCallback(fakeQueryInterface)
	ref := syscall.NewCallback(fakeRefCount)

	fakeEnumVtbl = &ole.IEnumVARIANTVtbl{
		IUnknownVtbl: ole.IUnknownVtbl{QueryInterface: qi, AddRef: ref, Release: ref},
		Next:         syscall.NewCallback(fakeEnumNext),
	}
	fakeEnumObj = &fakeCOMObject{vtbl: unsafe.Pointer(fakeEnumVtbl)}
	fakeEnumShared = uintptr(unsafe.Pointer(fakeEnumObj))

	fakeDispVtbl = &ole.IDispatchVtbl{
		IUnknownVtbl:  ole.IUnknownVtbl{QueryInterface: qi, AddRef: ref, Release: ref},
		GetIDsOfNames: syscall.NewCallback(fakeGetIDsOfNames),
		Invoke:        syscall.NewCallback(fakeInvokeNewEnum),
	}
	fakeDispObj = &fakeCOMObject{vtbl: unsafe.Pointer(fakeDispVtbl)}
}

// ForEach has always handed f the same *VARIANT on every call, overwritten by
// each IEnumVARIANT::Next. A caller that retains the pointer depends on that,
// so it is pinned here against a fake enumerator. The same pin against the WMI
// process list is TestForEach_PassesTheSameVariantToEveryCall_WMI (livecom).
func TestForEach_PassesTheSameVariantToEveryCall(t *testing.T) {
	fakeEnumObj.next = 0
	disp := (*ole.IDispatch)(unsafe.Pointer(fakeDispObj))

	var seen []*ole.VARIANT
	var values []int64
	err := ForEach(disp, func(v *ole.VARIANT) error {
		seen = append(seen, v)
		values = append(values, v.Val)
		return nil
	})
	require.NoError(t, err)

	require.Len(t, seen, fakeEnumItems)
	assert.Equal(t, []int64{1, 2, 3}, values)
	for i, p := range seen {
		assert.Same(t, seen[0], p, "call %d got a different *VARIANT", i)
	}
}
