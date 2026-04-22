//go:build !darwin
package stream

func HasScreenCaptureAccess() bool {
	return true
}

func RequestScreenCaptureAccess() {
}
