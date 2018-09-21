/* vim: set ts=2 sw=2 sts=2 et: */
package main

import (
  "net/http"
  "fmt"
//  "log"
//  "strings"
//  "time"
//  "sync"
  _ "net/http/pprof"
  "github.com/sirupsen/logrus"
  "github.com/takama/router"
  "github.com/skillcoder/homer/version"
  "github.com/skillcoder/homer/shutdown"
  infoHandler "github.com/skillcoder/go-common-handlers/info"
)

var log = logrus.New()

func init() {
  log.SetLevel(logrus.DebugLevel)
}

func main() {
  //log.Out = os.Stdout
  fmt.Printf("Homer v%s [%s]\n", version.RELEASE, version.BUILD)
  fmt.Printf("WWW: %s (%s)\n\n", version.REPO, version.COMMIT)

  config_load()

  fmt.Printf("Switch HOMER_MODE to %s\n", config.Mode)
  if config.Mode == "production" {
    logrus.SetFormatter(&logrus.JSONFormatter{})
    log.SetLevel(logrus.WarnLevel)
  } else {
    // The TextFormatter is default, you don't actually have to do this.
    //log.SetFormatter(&log.TextFormatter{})
    log.SetLevel(logrus.DebugLevel)
  }

  // Output to stdout instead of the default stderr
  // Can be any io.Writer, see below for File example
  //log.SetOutput(os.Stdout)
/*
  log.WithFields(logrus.Fields{
    "mode": config.Mode,
  }).Warn("Inited")
  */

  mqttConnect(config.Mqtt.Host, config.Mqtt.Port, config.Mqtt.Name)

  r := router.New()
  r.Logger = logger
  r.GET("/", home)

  // Readiness and liveness probes for Kubernetes
  r.GET("/info", infoHandler.Handler(version.RELEASE, version.REPO, version.COMMIT, version.BUILD))
  r.GET("/health", func(c *router.Control) {
    c.Code(http.StatusOK).Body(http.StatusText(http.StatusOK))
  })

  go r.Listen(config.Listen)

  if config.Mode == "development" {
    go func() {
      log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
  }

  clickhouse_connect()

  go dbLoop(5000)

  logger := log.WithField("event", "shutdown")
  sdHandler := shutdown.NewHandler(logger)
  sdHandler.RegisterShutdown(sd)
}

// sd does graceful dhutdown of the service
func sd() (string, error) {
  // if service has to finish some tasks before shutting down, these tasks must be finished her
  dbShutdown()
  // TODO(developer): wait for all gorutined ends
  // http://devs.cloudimmunity.com/gotchas-and-common-mistakes-in-go-golang/index.html#gor_app_exit
  return "Ok", nil
}
