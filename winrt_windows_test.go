//go:build windows

package ole

import (
	"errors"
	"runtime"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// roInitMultithreaded is RO_INIT_MULTITHREADED from roapi.h.
const roInitMultithreaded = 1

const (
	sFalse          = 0x00000001 // S_FALSE: already initialized in this mode
	rpcEChangedMode = 0x80010106 // RPC_E_CHANGED_MODE: thread already in another apartment
)

// roInitThread pins the test to one OS thread and initializes WinRT on it.
// Earlier tests can leave a reused thread in a single-threaded apartment;
// activation works there too, so that is accepted rather than failed on.
func roInitThread(t *testing.T) {
	t.Helper()
	runtime.LockOSThread()
	t.Cleanup(runtime.UnlockOSThread)

	err := RoInitialize(roInitMultithreaded)
	var oleErr *OleError
	switch {
	case err == nil, errors.As(err, &oleErr) && oleErr.Code() == sFalse:
		t.Cleanup(func() { RoUninitialize() })
	case errors.As(err, &oleErr) && oleErr.Code() == rpcEChangedMode:
		// not ours to uninitialize
	default:
		require.NoError(t, err)
	}
}

func TestHStringRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"ascii", "hello", "hello"},
		{"latin accents", "héllo wörld", "héllo wörld"},
		{"cjk", "日本語", "日本語"},
		// NewHString passes the rune count as the UTF-16 length, so every
		// surrogate pair costs one trailing code unit. Pinned, not endorsed.
		{"outside the BMP is truncated", "a😀b", "a😀"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := NewHString(tt.in)
			require.NoError(t, err)
			defer DeleteHString(h)

			assert.Equal(t, tt.want, h.String())
		})
	}
}

func TestRoActivateInstance_JsonObject(t *testing.T) {
	roInitThread(t)

	ins, err := RoActivateInstance("Windows.Data.Json.JsonObject")
	require.NoError(t, err)
	require.NotNil(t, ins)
	defer ins.Release()

	name, err := ins.GetRuntimeClassName()
	require.NoError(t, err)
	assert.Equal(t, "Windows.Data.Json.JsonObject", name)

	iids, err := ins.GetIids()
	require.NoError(t, err)
	require.NotEmpty(t, iids)
	for _, iid := range iids {
		assert.NotNil(t, iid)
		assert.False(t, IsEqualGUID(iid, &GUID{}), "GetIids returned an all-zero IID")
	}

	level, err := ins.GetTrustLevel()
	require.NoError(t, err)
	assert.Equal(t, uint32(0), level, "JsonObject is BaseTrust")
}

func TestRoGetActivationFactory_JsonObject(t *testing.T) {
	roInitThread(t)

	factory, err := RoGetActivationFactory("Windows.Data.Json.JsonObject", IID_IInspectable)
	require.NoError(t, err)
	require.NotNil(t, factory)
	defer factory.Release()

	iids, err := factory.GetIids()
	require.NoError(t, err)
	assert.NotEmpty(t, iids)
}

// GetDimensions and GetSize hand back the UINT that SafeArrayGetDim and
// SafeArrayGetElemsize return, converted to a *uint32 rather than pointed to.
// This pins that behaviour; dereferencing either result faults.
func TestSafeArrayConversion_DimensionsAndSizeAreIntegersInPointers(t *testing.T) {
	array, err := safeArrayCreateVector(VT_I4, 0, 3)
	require.NoError(t, err)
	require.NotNil(t, array)
	defer safeArrayDestroy(array)

	sac := &SafeArrayConversion{array}

	dims, err := sac.GetDimensions()
	assert.Equal(t, uintptr(1), uintptr(unsafe.Pointer(dims)), "dimension count carried in the pointer value")
	_ = err // the error is syscall's last-error, not an HRESULT; not part of what is pinned here

	size, err := sac.GetSize()
	assert.Equal(t, uintptr(4), uintptr(unsafe.Pointer(size)), "VT_I4 element size carried in the pointer value")
	_ = err // as above
}
