/* vim: set ts=2 sw=2 sts=2 et: */
package main

import (
  "net/http"
  "fmt"
//  "log"
  "os"
  "strings"
//  "time"
//  "sync"
  _ "net/http/pprof"
  "github.com/sirupsen/logrus"
  "github.com/takama/router"
  "github.com/micro/mdns"
  MQTT "github.com/eclipse/paho.mqtt.golang"
  "github.com/skillcoder/homer/version"
  "github.com/skillcoder/homer/shutdown"
  infoHandler "github.com/skillcoder/go-common-handlers/info"
)

var SERVICE_MODE = "production";

var espKnownNodes map[string]bool = make(map[string]bool)

var log = logrus.New()

func init() {
  // set environment depending on an environment variable or command-line flag
  if os.Getenv("SERVICE_MODE") != "" {
    SERVICE_MODE = os.Getenv("SERVICE_MODE")
  }

  if SERVICE_MODE == "production" {
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
    "mode": SERVICE_MODE,
  }).Warn("Inited")
  */
}

func main() {
  //log.Out = os.Stdout
  fmt.Printf("Homer v%s [%s] (%s)\n", version.RELEASE, version.BUILD, SERVICE_MODE)
  fmt.Printf("WWW: %s (%s)\n\n", version.REPO, version.COMMIT)

  config_load()

  http_listen := os.Getenv("HOMER_SERVICE_LISTEN")
  if len(http_listen) == 0 {
    log.Fatal("Required env parameter HOMER_SERVICE_LISTEN [ip:port] is not set")
  }

  service_name := os.Getenv("HOMER_SERVICE_NAME")
  if len(service_name) == 0 {
      log.Fatal("Required env parameter HOMER_SERVICE_NAME [go-homer-server] is not set")
  }

  //create a ClientOptions struct setting the broker address, clientid, turn
  //off trace output and set the default message handler
  opts := MQTT.NewClientOptions().AddBroker("tcp://"+config.Mqtt.Host+':'+string(config.Mqtt.Port))
  opts.SetClientID(service_name)
  opts.SetDefaultPublishHandler(mqttMessageHandler)

  //create and start a client using the above ClientOptions
  mqttClient := MQTT.NewClient(opts)
  if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
    log.Panic(token.Error())
    panic(token.Error())
  }

  //subscribe to the topic /go-mqtt/sample and request messages to be delivered
  //at a maximum qos of zero, wait for the receipt to confirm the subscription
  if token := mqttClient.Subscribe("/esp/init", 0, nil); token.Wait() && token.Error() != nil {
    log.Fatal(token.Error())
    os.Exit(1)
  }

  //Publish 5 messages to /go-mqtt/sample at qos 1 and wait for the receipt
  //from the server after sending each message
  /*
  for i := 0; i < 5; i++ {
    text := fmt.Sprintf("this is msg #%d!", i)
    token := mqttClient.Publish("go-mqtt/sample", 0, false, text)
    token.Wait()
  }

  time.Sleep(3 * time.Second)
  */

  //unsubscribe from /go-mqtt/sample
  /*
  if token := mqttClient.Unsubscribe("go-mqtt/sample"); token.Wait() && token.Error() != nil {
    fmt.Println(token.Error())
    os.Exit(1)
  }
  */
  //mqttClient.Disconnect(250)

  // Make a channel for results and start listening
  entriesCh := make(chan *mdns.ServiceEntry, 8)
  go func() {
    for entry := range entriesCh {
        //fmt.Printf("Got new entry: %v\n", entry)
      log.Infof("New node detected: (%s) [%s]", entry.Host, entry.AddrV4)
      s := strings.SplitN(entry.Host, ".", 2)
      mqttNodeSubscribe(mqttClient, s[0])
    }

    log.Info("Node discovery thread ended");
  }()

  // Start the lookup
  err := mdns.Lookup("_homer._tcp", entriesCh)
  if err != nil {
    log.Error(err)
  }

  close(entriesCh)

  r := router.New()
  r.Logger = logger
  r.GET("/", home)

  // Readiness and liveness probes for Kubernetes
  r.GET("/info", infoHandler.Handler(version.RELEASE, version.REPO, version.COMMIT, version.BUILD))
  r.GET("/health", func(c *router.Control) {
    c.Code(http.StatusOK).Body(http.StatusText(http.StatusOK))
  })

  go r.Listen(http_listen)

  if SERVICE_MODE == "development" {
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
