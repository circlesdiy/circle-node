package version

import "runtime"

const (
	// Version is the application version (semantic versioning)
	Version = "0.1.0-alpha"

	// AppName is the application name
	AppName = "Circles"
)

// GoVersion returns the Go runtime version
func GoVersion() string {
	return runtime.Version()
}
