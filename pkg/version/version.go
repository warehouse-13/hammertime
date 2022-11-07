package version

// PackageName is the name of the package, all commands fall under this name.
const PackageName = "hammertime"

var (
	Version    = "undefined" // Specifies the cli version
	BuildDate  = "undefined" // Specifies the build date
	CommitHash = "undefined" // Specifies the git commit hash
)
