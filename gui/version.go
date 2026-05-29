package gui

// Version is the current version of the fynerisor library
const Version = "0.4.1"

// appVersion stores the embedding application's version
// Set via SetAppVersion() to enable version checking in require() calls
var appVersion string = ""

// VersionInfo provides detailed version information
type VersionInfo struct {
	Version     string // Fynerisor library version
	AppVersion  string // Embedding application version (if set)
	RisorCompat string // Compatible Risor version
	FyneCompat  string // Compatible Fyne version
}

// GetVersion returns version information for this fynerisor release
func GetVersion() VersionInfo {
	return VersionInfo{
		Version:     Version,
		AppVersion:  appVersion,
		RisorCompat: "v2.x",
		FyneCompat:  "v2.7+",
	}
}

// SetAppVersion sets the embedding application's version.
// This version is used when scripts call require(["vX.Y.Z"]) to check
// compatibility with the embedding application, not fynerisor itself.
//
// Example:
//
//	func main() {
//	    fynerisor.SetAppVersion("1.2.3")
//	    w := fynerisor.NewApp("My App")
//	    // Scripts can now use: require(["v1.2"])
//	    w.LoadScript(script)
//	    w.ShowAndRun()
//	}
//
// Scripts should use semantic versioning format: "v1.2.3" or "v1.2"
func SetAppVersion(version string) {
	appVersion = version
}

// GetAppVersion returns the embedding application's version.
// Returns empty string if not set.
func GetAppVersion() string {
	return appVersion
}
