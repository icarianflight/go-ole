//go:build windows && livecom

// Live test: queries state of the host it runs on. Excluded unless built with
// -tags livecom.

package oleutil

import (
	"runtime"
	"testing"

	ole "github.com/icarianflight/go-ole"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ForEach has always handed f the same *VARIANT on every call, overwritten by
// each IEnumVARIANT::Next. A caller that retains the pointer depends on that,
// so it is pinned here against the WMI process list. The non-live pin of the same
// behaviour is TestForEach_PassesTheSameVariantToEveryCall.
func TestForEach_PassesTheSameVariantToEveryCall_WMI(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	require.NoError(t, ole.CoInitialize(0))
	defer ole.CoUninitialize()

	unknown, err := CreateObject("WbemScripting.SWbemLocator")
	require.NoError(t, err)
	defer unknown.Release()

	locator, err := unknown.QueryInterface(ole.IID_IDispatch)
	require.NoError(t, err)
	defer locator.Release()

	services, err := CallMethod(locator, "ConnectServer")
	require.NoError(t, err)
	defer services.Clear()

	processes, err := CallMethod(services.ToIDispatch(), "ExecQuery", "SELECT * FROM Win32_Process")
	require.NoError(t, err)
	defer processes.Clear()

	var seen []*ole.VARIANT
	var names []string
	err = ForEach(processes.ToIDispatch(), func(v *ole.VARIANT) error {
		seen = append(seen, v)
		name, err := GetProperty(v.ToIDispatch(), "Name")
		if err != nil {
			return err
		}
		names = append(names, name.ToString())
		name.Clear()
		return v.Clear()
	})
	require.NoError(t, err)

	require.Greater(t, len(seen), 1, "need at least two processes to compare pointers")
	for i, p := range seen {
		assert.Same(t, seen[0], p, "call %d got a different *VARIANT", i)
	}
	assert.NotEmpty(t, names[0])
}
