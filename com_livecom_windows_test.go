//go:build windows && livecom

// Live tests: they resolve ProgIDs through the host's COM registrations.
// Excluded unless built with -tags livecom.

package ole

import "testing"

func TestClsidFromProgID_WindowsMediaNSSManager(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Fail()
		}
	}()

	expected := &GUID{0x92498132, 0x4D1A, 0x4297, [8]byte{0x9B, 0x78, 0x9E, 0x2E, 0x4B, 0xA9, 0x9C, 0x07}}

	coInitialize()
	defer CoUninitialize()
	actual, err := CLSIDFromProgID("WMPNSSCI.NSSManager")
	if err == nil {
		if !IsEqualGUID(expected, actual) {
			t.Log(err)
			t.Logf("Actual GUID: %+v\n", actual)
			t.Fail()
		}
	}
}

func TestCreateInstance_WindowsMediaNSSManager(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Fail()
		}
	}()

	expected := &GUID{0x92498132, 0x4D1A, 0x4297, [8]byte{0x9B, 0x78, 0x9E, 0x2E, 0x4B, 0xA9, 0x9C, 0x07}}

	coInitialize()
	defer CoUninitialize()
	actual, err := CLSIDFromProgID("WMPNSSCI.NSSManager")

	if err == nil {
		if !IsEqualGUID(expected, actual) {
			t.Log(err)
			t.Logf("Actual GUID: %+v\n", actual)
			t.Fail()
		}

		unknown, err := CreateInstance(actual, IID_IUnknown)
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		unknown.Release()
	}
}
