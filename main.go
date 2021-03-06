package main

/* vim: set ts=2 sw=2 sts=2 ff=unix ft=go noet: */

import (
	"fmt"
	"net/http"
	"os"
	//  "log"
	//  "strings"
	//  "time"
	//  "sync"
	_ "net/http/pprof"

	"github.com/sirupsen/logrus"
	infoHandler "github.com/skillcoder/go-common-handlers/info"
	"github.com/skillcoder/homer/shutdown"
	statHandler "github.com/skillcoder/homer/stat"
	"github.com/takama/router"
)

var log = logrus.New()

func init() {
	log.SetLevel(logrus.DebugLevel)
}

func check(e error) {
	if e != nil {
		failWith(e.Error())
	}
}

/*
func checkDefer(err error) {
	if err != nil {
		log.Error(err)
	}
}

func checkFunc(f func() error) {
	if err := f(); err != nil {
		fmt.Println("Error while defer:", err)
	}
}
*/

func failWith(msg string) {
	log.Fatal(msg)
	if _, err := fmt.Fprintln(os.Stderr, msg); err != nil {
		log.Error(err)
	}

	os.Exit(1)
}

func verbosePrint(msg string) {
	if config.Verbose {
		fmt.Println("[V]", msg)
	}
}

func main() {
	//log.Out = os.Stdout
	fmt.Printf("Homer v%s [%s]\n", versionRELEASE, versionBUILD)
	fmt.Printf("WWW: %s (%s)\n\n", versionREPO, versionCOMMIT)

	configLoad()

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

	mqttConnect(config.Mqtt.Host, config.Mqtt.Port, config.Mqtt.Name, config.Mqtt.User, config.Mqtt.Pass)

	r := router.New()
	r.Logger = handlersLogger
	r.GET("/", handlersHome)

	// Readiness and liveness probes for Kubernetes
	r.GET("/info", infoHandler.Handler(versionRELEASE, versionREPO, versionCOMMIT, versionBUILD))
	r.GET("/stat", statHandler.Handler(dbGetStatHandler))
	r.GET("/health", func(c *router.Control) {
		c.Code(http.StatusOK).Body(http.StatusText(http.StatusOK))
	})

	go r.Listen(config.Listen)

	if config.Mode == "development" {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	clickhouseConnect()
	dbInit()

	go dbLoop(int(config.AggregatePeriod))

	logrus.RegisterExitHandler(func() {
		// gracefully shutdown something...
	})
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
