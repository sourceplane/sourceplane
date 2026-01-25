package version

// Version information
var (
	// Version is the current version of the CLI
	Version = "0.2.1"

	// GitCommit is the git commit hash (set during build)
	GitCommit = "unknown"

	// BuildDate is the date the binary was built (set during build)
	BuildDate = "unknown"
)

// GetVersion returns the full version string
func GetVersion() string {
	if GitCommit != "unknown" {
		return Version + " (" + GitCommit + ")"
	}
	return Version
}

// GetFullVersion returns version with all build information
func GetFullVersion() string {
	return "Version: " + Version + "\nCommit: " + GitCommit + "\nBuilt: " + BuildDate
}
