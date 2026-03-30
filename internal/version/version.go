package version

// Build-time variables injected via -ldflags.
// Defaults are used when building without the Makefile (e.g. go run).
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// Info holds the build metadata.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
}

// Get returns the current build information.
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
	}
}
