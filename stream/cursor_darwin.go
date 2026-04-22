//go:build darwin
package stream

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework CoreGraphics
#import <Cocoa/Cocoa.h>
#import <CoreGraphics/CoreGraphics.h>

void getMacMousePosition(float *x, float *y) {
    NSPoint loc = [NSEvent mouseLocation];
    CGFloat mainHeight = CGDisplayBounds(CGMainDisplayID()).size.height;
    *x = loc.x;
    *y = mainHeight - loc.y;
}
*/
import "C"

func GetCursorPos() (float64, float64) {
	var x, y C.float
	C.getMacMousePosition(&x, &y)
	return float64(x), float64(y)
}
