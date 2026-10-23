# Version 1.4.1 (icarianflight fork)

Bug fix release. This is a deliberate departure from upstream go-ole behaviour;
the exported API (names, signatures and types) is unchanged.

**Upgrade note:** a caller that worked around the old `GetDimensions()` /
`GetSize()` by reading the returned pointer's address as the value (for
example `uintptr(unsafe.Pointer(dims))`) must now dereference it (`*dims`).
`GetMessage` also now returns a nil error unless the call itself failed.

 * Fixed `SafeArrayConversion.GetDimensions()` and `GetSize()`. They returned
   the Win32 UINT result of `SafeArrayGetDim` / `SafeArrayGetElemsize`
   converted into a `*uint32`, so dereferencing the result read address 1 (or
   2, 4, 8...) and crashed. They now return a pointer to a real `uint32`
   holding the value.
 * The same two calls returned the syscall last-error as `err`, which is
   non-nil even on success ("The operation completed successfully."). `err` is
   now nil on success. These Win32 functions have no failure HRESULT and
   return 0 for a NULL or invalid array, and `SafeArrayGetElemsize` also
   returns 0 for a not-yet-initialised descriptor (one from
   `SafeArrayAllocDescriptor` before its element size is set); that case now
   returns an `*OleError` with code `E_INVALIDARG` (and a non-nil pointer to
   0).
 * Fixed `GetMessage` and `DispatchMessage` on 64-bit Windows. The exported
   `Msg` is the 28-byte 32-bit layout, but the 64-bit MSG is 48 bytes, so
   `GetMessageW` wrote 20 bytes past `Msg` and `Message` was read from the
   wrong offset. Both functions now use a correctly laid out internal MSG and
   copy to and from `Msg`. `GetMessage` returns an error only when the result
   is -1 (previously `err` was always non-nil). Limits that remain because the
   exported types cannot change in a patch release: `Msg.Hwnd`, `Wparam` and
   `Lparam` are 32 bits, so on 64-bit Windows a pointer-valued `Wparam` or
   `Lparam` is truncated; the `hwnd` filter argument is `uint32` (it is
   sign-extended, so `0xFFFFFFFF` is the thread-messages filter hWnd -1, and
   `Msg.Hwnd` is sign-extended likewise in `DispatchMessage`); the
   `DispatchMessage` result (an LRESULT) is truncated to `int32`.
 * `safeArrayCreate`, `safeArrayCreateEx`, `safeArrayCreateVector` and
   `safeArrayCreateVectorEx` now judge failure by a NULL return, not by the
   last-error string. NULL gives a nil array and an `E_INVALIDARG` error
   (these calls have no HRESULT, so this covers invalid arguments and out of
   memory alike); success gives a nil error even when a stale last-error was
   set.
 * `VariantInit` and `SysFreeString` no longer return spurious errors. The
   Win32 functions return void, so the old code was testing leftover register
   contents. They now always return nil.
 * Live tests (WMI queries, QuickBooks, Windows Media ProgID lookups, WinRT
   `JsonObject` activation, `TestSafeArrayConversionString`) are now behind
   the `livecom` build tag and no longer run by default; use
   `go test -tags livecom`. CI runs the WMI, Windows Media and WinRT
   activation ones with the tag.
   `TestForEach_PassesTheSameVariantToEveryCall` now pins the `ForEach`
   semantics against a hand-built in-process enumerator; the WMI version is
   kept as `..._WMI` behind the tag.
   `TestError` now uses `IIDFromString` on a malformed string, which is parsed
   without consulting the registry, instead of looking up a ProgID.
 * Replaced the test that pinned the old pointer-carries-the-integer behaviour
   with tests covering 1-D and 2-D arrays and element sizes 1, 4 and 8, and
   added tests for the other fixes above.

# Version 1.4.0 (icarianflight fork)

First release of the fork. See the README's "About this fork" section for its
relationship to upstream go-ole.

**Minimum supported version is now Go 1.27** (was Go 1.12). The module path is
`github.com/icarianflight/go-ole`; the exported API and behaviour are unchanged
from upstream go-ole master (as of 28687b4, October 2026).

 * Updated `golang.org/x/sys` to v0.49.0 (clears GO-2026-5024) and
   `github.com/stretchr/testify` to v1.12.1.
 * Converted all `// +build` constraints to `//go:build` form.
 * `go vet` is now clean on all platforms: `reflect.SliceHeader` uses replaced
   with `unsafe.Slice`, and uintptr-to-pointer conversions routed through a
   vetted helper with identical generated code.
 * Replaced deprecated `syscall.Syscall*` calls with `syscall.SyscallN`.
 * Preserved pre-Go-1.22 `oleutil.ForEach` semantics (the shared `*VARIANT`
   per iteration) against the loop-variable change, pinned by a test.
 * GitHub Actions workflows updated to Go 1.27/stable and full-module coverage.
 * AppVeyor configuration removed; CI is GitHub Actions only.

# Version 1.x.x

* **Add more test cases and reference new test COM server project.** (Placeholder for future additions)

# Version 1.2.0-alphaX

**Minimum supported version is now Go 1.4. Go 1.1 support is deprecated, but should still build.**

 * Added CI configuration for Travis-CI and AppVeyor.
 * Added test InterfaceID and ClassID for the COM Test Server project.
 * Added more inline documentation (#83).
 * Added IEnumVARIANT implementation (#88).
 * Added IEnumVARIANT test cases (#99, #100, #101).
 * Added support for retrieving `time.Time` from VARIANT (#92).
 * Added test case for IUnknown (#64).
 * Added test case for IDispatch (#64).
 * Added test cases for scalar variants (#64, #76).

# Version 1.1.1

 * Fixes for Linux build.
 * Fixes for Windows build.

# Version 1.1.0

The change to provide building on all platforms is a new feature. The increase in minor version reflects that and allows those who wish to stay on 1.0.x to continue to do so. Support for 1.0.x will be limited to bug fixes.

 * Move GUID out of variables.go into its own file to make new documentation available.
 * Move OleError out of ole.go into its own file to make new documentation available.
 * Add documentation to utility functions.
 * Add documentation to variant receiver functions.
 * Add documentation to ole structures.
 * Make variant available to other systems outside of Windows.
 * Make OLE structures available to other systems outside of Windows.

## New Features

 * Library should now be built on all platforms supported by Go. Library will NOOP on any platform that is not Windows.
 * More functions are now documented and available on godoc.org.

# Version 1.0.1

 1. Fix package references from repository location change.

# Version 1.0.0

This version is stable enough for use. The COM API is still incomplete, but provides enough functionality for accessing COM servers using IDispatch interface.

There is no changelog for this version. Check commits for history.
