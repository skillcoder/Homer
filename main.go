package main

import (
	"net/http"
	"fmt"
//	"log"
	"os"
//	"time"
	"github.com/sirupsen/logrus"
	"github.com/takama/router"
	MQTT "github.com/eclipse/paho.mqtt.golang"
//	"github.com/skillcoder/homer/version"
	"github.com/skillcoder/homer/shutdown"
)

var log = logrus.New()

//define a function for the default message handler
var mqttMessageHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
  fmt.Printf("TOPIC: %s\n", msg.Topic())
  fmt.Printf("MSG: %s\n", msg.Payload())
}

// Run server: go build; env SERVICE_PORT=8000 step-by-step
// Try requests: curl http://127.0.0.1:8000/test
func main() {
  http_listen := os.Getenv("SERVICE_LISTEN")
  if len(http_listen) == 0 {
	  log.Fatal("Required env parameter SERVICE_LISTEN [ip:port] is not set")
  }

  mqtt_server := os.Getenv("MQTT_SERVER")
  if len(mqtt_server) == 0 {
	  log.Fatal("Required env parameter MQTT_SERVER [ip:port] is not set")
  }

  //create a ClientOptions struct setting the broker address, clientid, turn
  //off trace output and set the default message handler
  opts := MQTT.NewClientOptions().AddBroker("tcp://"+mqtt_server)
  opts.SetClientID("go-homer-server")
  opts.SetDefaultPublishHandler(mqttMessageHandler)

  //create and start a client using the above ClientOptions
  mqttClient := MQTT.NewClient(opts)
  if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
    panic(token.Error())
  }

  //subscribe to the topic /go-mqtt/sample and request messages to be delivered
  //at a maximum qos of zero, wait for the receipt to confirm the subscription
  if token := mqttClient.Subscribe("/esp/init", 0, nil); token.Wait() && token.Error() != nil {
    fmt.Println(token.Error())
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

  r := router.New()
  r.Logger = logger
  r.GET("/", home)

  // Readiness and liveness probes for Kubernetes
  r.GET("/info", func(c *router.Control) {
	  //common_handlers.Info(c, version.RELEASE, version.REPO, version.COMMIT)
  })
  r.GET("/healthz", func(c *router.Control) {
	  c.Code(http.StatusOK).Body(http.StatusText(http.StatusOK))
  })

  go r.Listen(http_listen)

  logger := log.WithField("event", "shutdown")
  sdHandler := shutdown.NewHandler(logger)
  sdHandler.RegisterShutdown(sd)
}

// sd does graceful dhutdown of the service
func sd() (string, error) {
	// if service has to finish some tasks before shutting down, these tasks must be finished her
	return "Ok", nil
}
