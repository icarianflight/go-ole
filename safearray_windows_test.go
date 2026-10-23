//go:build windows

package ole

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_safeArrayCreate(t *testing.T) {
	type args struct {
		variantType VT
		dimensions  uint32
		bounds      *SafeArrayBound
	}
	type want struct {
		Dimensions   uint16
		ElementsSize uint32
		LocksAmount  uint32
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "Create SafeArray",
			args: args{
				variantType: VT_UI1,
				dimensions:  uint32(1),
				bounds: &SafeArrayBound{
					Elements:   uint32(10000),
					LowerBound: int32(0),
				},
			},
			want: want{
				Dimensions:   uint16(1),
				ElementsSize: uint32(1),
				LocksAmount:  uint32(0),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := safeArrayCreate(tt.args.variantType, tt.args.dimensions, tt.args.bounds)
			if (err != nil) != tt.wantErr {
				t.Errorf("safeArrayCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, tt.want.Dimensions, got.Dimensions) {
				t.Errorf("safeArrayCreate() Dimensions not equal: got.Dimensions = %v, tt.want.Dimensions %v", got.Dimensions, tt.want.Dimensions)
			}
			if !assert.Equal(t, tt.want.ElementsSize, got.ElementsSize) {
				t.Errorf("safeArrayCreate() ElementsSize not equal: got.ElementsSize = %v, tt.want.ElementsSize %v", got.ElementsSize, tt.want.ElementsSize)
			}
			if !assert.Equal(t, tt.want.LocksAmount, got.LocksAmount) {
				t.Errorf("safeArrayCreate() LocksAmount not equal: got.LocksAmount = %v, tt.want.LocksAmount %v", got.LocksAmount, tt.want.LocksAmount)
			}
			safeArrayDestroy(got)
		})
	}
}

func TestSafeArrayConversion_GetDimensionsAndGetSize(t *testing.T) {
	ptrSize := uint32(unsafe.Sizeof(uintptr(0)))

	tests := []struct {
		name     string
		create   func() (*SafeArray, error)
		wantDims uint32
		wantSize uint32
	}{
		{
			name:     "1-D VT_UI1",
			create:   func() (*SafeArray, error) { return safeArrayCreateVector(VT_UI1, 0, 5) },
			wantDims: 1,
			wantSize: 1,
		},
		{
			name:     "1-D VT_I4",
			create:   func() (*SafeArray, error) { return safeArrayCreateVector(VT_I4, 0, 3) },
			wantDims: 1,
			wantSize: 4,
		},
		{
			name:     "1-D VT_R8",
			create:   func() (*SafeArray, error) { return safeArrayCreateVector(VT_R8, 0, 3) },
			wantDims: 1,
			wantSize: 8,
		},
		{
			name:     "1-D VT_BSTR is pointer sized",
			create:   func() (*SafeArray, error) { return safeArrayCreateVector(VT_BSTR, 0, 2) },
			wantDims: 1,
			wantSize: ptrSize,
		},
		{
			name: "2-D VT_UI1",
			create: func() (*SafeArray, error) {
				b := []SafeArrayBound{{Elements: 2}, {Elements: 3}}
				return safeArrayCreate(VT_UI1, 2, &b[0])
			},
			wantDims: 2,
			wantSize: 1,
		},
		{
			name: "2-D VT_I4",
			create: func() (*SafeArray, error) {
				b := []SafeArrayBound{{Elements: 2}, {Elements: 3}}
				return safeArrayCreate(VT_I4, 2, &b[0])
			},
			wantDims: 2,
			wantSize: 4,
		},
		{
			name: "2-D VT_R8",
			create: func() (*SafeArray, error) {
				b := []SafeArrayBound{{Elements: 2}, {Elements: 3}}
				return safeArrayCreate(VT_R8, 2, &b[0])
			},
			wantDims: 2,
			wantSize: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			array, err := tt.create()
			require.NoError(t, err)
			require.NotNil(t, array)
			defer safeArrayDestroy(array)

			sac := &SafeArrayConversion{array}

			dims, err := sac.GetDimensions()
			require.NoError(t, err)
			require.NotNil(t, dims)
			assert.Equal(t, tt.wantDims, *dims)

			size, err := sac.GetSize()
			require.NoError(t, err)
			require.NotNil(t, size)
			assert.Equal(t, tt.wantSize, *size)
		})
	}
}

// The Win32 calls have no failure HRESULT; they return 0 for a NULL array.
// That is reported as E_INVALIDARG, with a non-nil pointer to 0 so a caller
// that ignores the error still cannot fault.
func TestSafeArrayConversion_GetDimensionsAndGetSize_NullArray(t *testing.T) {
	sac := &SafeArrayConversion{nil}

	dims, err := sac.GetDimensions()
	var oleErr *OleError
	require.ErrorAs(t, err, &oleErr)
	assert.Equal(t, uintptr(E_INVALIDARG), oleErr.Code())
	require.NotNil(t, dims)
	assert.Equal(t, uint32(0), *dims)

	size, err := sac.GetSize()
	require.ErrorAs(t, err, &oleErr)
	assert.Equal(t, uintptr(E_INVALIDARG), oleErr.Code())
	require.NotNil(t, size)
	assert.Equal(t, uint32(0), *size)
}

func TestSafeArrayCreate_ReportsFailureByNullReturn(t *testing.T) {
	bounds := SafeArrayBound{Elements: 3}

	failing := []struct {
		name   string
		create func() (*SafeArray, error)
	}{
		{"Create with zero dimensions", func() (*SafeArray, error) { return safeArrayCreate(VT_I4, 0, &bounds) }},
		{"CreateEx with zero dimensions", func() (*SafeArray, error) { return safeArrayCreateEx(VT_I4, 0, &bounds, 0) }},
		{"Create with VT_EMPTY", func() (*SafeArray, error) { return safeArrayCreate(VT_EMPTY, 1, &bounds) }},
		{"CreateEx with VT_EMPTY", func() (*SafeArray, error) { return safeArrayCreateEx(VT_EMPTY, 1, &bounds, 0) }},
		{"CreateVector with VT_EMPTY", func() (*SafeArray, error) { return safeArrayCreateVector(VT_EMPTY, 0, 3) }},
		{"CreateVectorEx with VT_EMPTY", func() (*SafeArray, error) { return safeArrayCreateVectorEx(VT_EMPTY, 0, 3, 0) }},
	}
	for _, tt := range failing {
		t.Run(tt.name, func(t *testing.T) {
			array, err := tt.create()
			assert.Nil(t, array)
			assert.Error(t, err)
		})
	}

	succeeding := []struct {
		name   string
		create func() (*SafeArray, error)
	}{
		{"Create", func() (*SafeArray, error) { return safeArrayCreate(VT_I4, 1, &bounds) }},
		{"CreateEx", func() (*SafeArray, error) { return safeArrayCreateEx(VT_I4, 1, &bounds, 0) }},
		{"CreateVector", func() (*SafeArray, error) { return safeArrayCreateVector(VT_I4, 0, 3) }},
		{"CreateVectorEx", func() (*SafeArray, error) { return safeArrayCreateVectorEx(VT_I4, 0, 3, 0) }},
	}
	for _, tt := range succeeding {
		t.Run(tt.name, func(t *testing.T) {
			array, err := tt.create()
			require.NoError(t, err)
			require.NotNil(t, array)
			safeArrayDestroy(array)
		})
	}
}
