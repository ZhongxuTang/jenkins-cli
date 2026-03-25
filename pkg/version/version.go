package version

var (
	// Version is the semantic version for the CLI binary; default "dev" for local builds.
	Version = "dev"
	// Commit is the Git SHA corresponding to the build.
	Commit = "none"
	// BuildDate holds the RFC3339 timestamp of the build.
	BuildDate = "unknown"
)

// Info returns the current version metadata as a map for easy inspection or formatting.
func Info() map[string]string {
	return map[string]string{
		"version":   Version,
		"commit":    Commit,
		"buildDate": BuildDate,
	}
}
