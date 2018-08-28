package version

// Rewrite by
// go build -ldflags "-X github.com/skillcoder/homer/version.BUILD=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X github.com/skillcoder/homer/version.COMMIT=`git rev-parse HEAD` -X github.com/skillcoder/homer/version.RELEASE='0.0.0'"
// see how to use: https://github.com/golang/go/wiki/GcToolchainTricks
var (
	// REPO returns the git repository URL
	REPO string = "https://github.com/skillcoder/Homer.git"
	// RELEASE returns the release version
	RELEASE string = "UNKNOWN"
	// COMMIT returns the short sha from git
	COMMIT string = "UNKNOWN"
	// STAMP returns time and date of build
	BUILD string = "UNKNOWN"
)
