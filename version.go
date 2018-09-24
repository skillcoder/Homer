package main

/* vim: set ts=2 sw=2 sts=2 ff=unix ft=go noet: */

// Rewrite by -ldflags
// see how to use: https://github.com/golang/go/wiki/GcToolchainTricks
// or in Makefile
var (
	// REPO returns the git repository URL
	versionREPO = "https://github.com/skillcoder/Homer.git"
	// RELEASE returns the release version
	versionRELEASE = "UNKNOWN"
	// COMMIT returns the short sha from git
	versionCOMMIT = "UNKNOWN"
	// STAMP returns time and date of build
	versionBUILD = "UNKNOWN"
)
