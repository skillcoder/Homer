package version

// Rewrite by
// go build -ldflags "-X github.com/skillcoder/homer/version.BUILD=`date -u '+%Y-%m-%d_%H:%M:%S%p'` -X github.com/skillcoder/homer/version.COMMIT=`git rev-parse HEAD` -X github.com/skillcoder/homer/version.RELEASE=`cat VERSION`"
// see how to use: https://github.com/golang/go/wiki/GcToolchainTricks
var (
	// REPO returns the git repository URL
	REPO = "https://github.com/skillcoder/Homer.git"
	// RELEASE returns the release version
	RELEASE = "UNKNOWN"
	// COMMIT returns the short sha from git
	COMMIT = "UNKNOWN"
	// STAMP returns time and date of build
	BUILD = "UNKNOWN"
)
