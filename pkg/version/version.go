package version

// Version is injected at build time via ldflags.
var Version = "dev"

// String returns the version string.
func String() string {
	return Version
}
