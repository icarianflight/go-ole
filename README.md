# Go OLE

Go bindings for Windows COM using shared libraries instead of cgo.

By Yasuhiro Matsumoto.

## Install

To experiment with go-ole, you can just compile and run the example program:

```
go get github.com/icarianflight/go-ole
cd /path/to/go-ole/
go test

cd /path/to/go-ole/_example/excel
go run excel.go
```

## Testing

1. Download a release from https://github.com/go-ole/test-com-server
2. Register the COM server
   - On Windows 32-bit: 
     
     `c:\Windows\Microsoft.NET\Framework\v4.0.30319\RegAsm.exe /codebase /nologo c:\path\to\TestCOMServer.dll`
   - On Windows 64-bit:
   
     `c:\Windows\Microsoft.NET\Framework64\v4.0.30319\RegAsm.exe /codebase /nologo c:\path\to\TestCOMServer.dll`
3. `go test`

## Multithreading

You have two solutions for handling gothreads and multithreading.

1. You may lock the function or gothread to a single thread.
    ```go
    runtime.LockOSThread()
    defer runtime.UnlockOSThread()
    ```
2. Use [scjalliance/comshim](https://github.com/scjalliance/comshim)

The key to any solution is that you must call `CoUninitialize()` for every `CoInitialize()`.

## Continuous Integration

Continuous integration runs on GitHub Actions (`.github/workflows/`). The amd64 and 386 test jobs each download the [test COM server](https://github.com/go-ole/test-com-server) (pinned to v1.0.2, the release matching these tests), register it with `RegAsm`, and then run `go build`, `go vet` and `go test` on the same Windows runner, with Go 1.27 and the current stable release. A separate job runs `go test -race` on amd64. arm64 is build and vet only, because arm64 binaries cannot run on the x64 runner. The library cannot be tested on Linux, since the COM API is specific to Windows.

A green run on all Windows test legs is the bar for every commit.

## Versioning

Go OLE uses [semantic versioning](http://semver.org) for version numbers, which is similar to the version contract of the Go language. Which means that the major version will always maintain backwards compatibility with minor versions. Minor versions will only add new additions and changes. Fixes will always be in patch. 

This contract should allow you to upgrade to new minor and patch versions without breakage or modifications to your existing code. Leave a ticket, if there is breakage, so that it could be fixed.

## LICENSE

Under the MIT License: http://mattn.mit-license.org/2013
