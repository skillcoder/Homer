package main
/* vim: set ts=2 sw=2 sts=2 et: */

import (
  "fmt"

  "github.com/takama/router"
)

// home returns the path of current request
func home(c *router.Control) {
	if _, err := fmt.Fprintf(c.Writer, "Repo: %s, Commit: %s, Version: %s, Build: %s",
    versionREPO, versionCOMMIT, versionRELEASE, versionBUILD); err != nil {
    log.Error("Cant write [home]:", err)
  }
}

// logger provides a log of requests
func logger(c *router.Control) {
	remoteAddr := c.Request.Header.Get("X-Forwarded-For")
	if remoteAddr == "" {
		remoteAddr = c.Request.RemoteAddr
	}

	log.Infof("%s %s %s", remoteAddr, c.Request.Method, c.Request.URL.Path)
}
