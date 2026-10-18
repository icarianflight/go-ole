package ole

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

// These tests pin how VARIANT reinterprets its integer Val slot as a COM
// pointer. The pointees are Go-allocated stand-ins for COM objects; they are
// only compared by address, never called through.

func TestVariantToIUnknown(t *testing.T) {
	obj := new(IUnknown)
	addr := int64(uintptr(unsafe.Pointer(obj)))

	tests := []struct {
		name string
		v    VARIANT
		want *IUnknown
	}{
		{"unknown returns the pointer held in Val", VARIANT{VT: VT_UNKNOWN, Val: addr}, obj},
		{"unknown with zero Val returns nil", VARIANT{VT: VT_UNKNOWN, Val: 0}, nil},
		{"other type returns nil", VARIANT{VT: VT_DISPATCH, Val: addr}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Same(t, tt.want, tt.v.ToIUnknown())
		})
	}
}

func TestVariantToIDispatch(t *testing.T) {
	obj := new(IDispatch)
	addr := int64(uintptr(unsafe.Pointer(obj)))

	tests := []struct {
		name string
		v    VARIANT
		want *IDispatch
	}{
		{"dispatch returns the pointer held in Val", VARIANT{VT: VT_DISPATCH, Val: addr}, obj},
		{"dispatch with zero Val returns nil", VARIANT{VT: VT_DISPATCH, Val: 0}, nil},
		{"other type returns nil", VARIANT{VT: VT_UNKNOWN, Val: addr}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Same(t, tt.want, tt.v.ToIDispatch())
		})
	}
}

func TestVariantToArray(t *testing.T) {
	arr := new(SafeArray)
	addr := int64(uintptr(unsafe.Pointer(arr)))

	tests := []struct {
		name string
		v    VARIANT
		want *SafeArray
	}{
		{"VT_SAFEARRAY wraps the pointer held in Val", VARIANT{VT: VT_SAFEARRAY, Val: addr}, arr},
		{"VT_ARRAY flag wraps the pointer held in Val", VARIANT{VT: VT_ARRAY | VT_VARIANT, Val: addr}, arr},
		{"array with zero Val wraps nil", VARIANT{VT: VT_ARRAY | VT_BSTR, Val: 0}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v.ToArray()
			if assert.NotNil(t, got) {
				assert.Same(t, tt.want, got.Array)
			}
		})
	}

	t.Run("non-array type returns nil", func(t *testing.T) {
		v := VARIANT{VT: VT_I4, Val: addr}
		assert.Nil(t, v.ToArray())
	})
}
