package main

import (
	"fmt"

    "github.com/skillcoder/homer/version"
	"github.com/takama/router"
)

// home returns the path of current request
func home(c *router.Control) {
	fmt.Fprintf(c.Writer, "Repo: %s, Commit: %s, Version: %s, Build: %s", version.REPO, version.COMMIT, version.RELEASE, version.BUILD)
}

// logger provides a log of requests
func logger(c *router.Control) {
	remoteAddr := c.Request.Header.Get("X-Forwarded-For")
	if remoteAddr == "" {
		remoteAddr = c.Request.RemoteAddr
	}
	log.Infof("%s %s %s", remoteAddr, c.Request.Method, c.Request.URL.Path)
}
