//go:build darwin
package stream

/*
#cgo LDFLAGS: -framework CoreGraphics
#include <CoreGraphics/CoreGraphics.h>

bool checkMacScreenCapturePermission() {
    if (__builtin_available(macOS 10.15, *)) {
        return CGPreflightScreenCaptureAccess();
    }
    return true;
}

bool requestMacScreenCapturePermission() {
    if (__builtin_available(macOS 10.15, *)) {
        return CGRequestScreenCaptureAccess();
    }
    return true;
}
*/
import "C"

func HasScreenCaptureAccess() bool {
	return bool(C.checkMacScreenCapturePermission())
}

func RequestScreenCaptureAccess() {
	C.requestMacScreenCapturePermission()
}
