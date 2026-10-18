# icarianflight fork

**Minimum supported version is now Go 1.27** (was Go 1.12). The module path is
`github.com/icarianflight/go-ole`; the exported API and behaviour are unchanged
from upstream go-ole master.

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
