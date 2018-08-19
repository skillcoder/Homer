package main

import (
	"net/http"
	"fmt"
//	"log"
	"os"
//	"time"
    "encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/takama/router"
	MQTT "github.com/eclipse/paho.mqtt.golang"
//	"github.com/skillcoder/homer/version"
	"github.com/skillcoder/homer/shutdown"
)

var log = logrus.New()

type mqttConnectedPacket struct {
    Act    string  `json:"act"`
    Time   uint32  `json:"time"`
    Name   string  `json:"name"`
	Ap     string  `json:"ap"`
	Mac    string  `json:"mac"`
	Ip     string  `json:"ip"`
	Rssi   int16   `json:"rssi"`
	Vcc    float32 `json:"vcc"`
	H      string  `json:"h"`
	V      string  `json:"v"`
	UPD    uint32  `json:"upd"`
	Inited uint32  `json:"inited"`
	Free   uint32  `json:"free"`
}

type mqttStatusPacket struct{
    Act    string  `json:"act"`
    Time   uint32  `json:"time"`
    Ap     string  `json:"ap"`
    Rssi   int16   `json:"rssi"`
    Vcc    float32 `json:"vcc"`
    Free   uint32  `json:"free"`
	Uptime uint32  `json:"uptime"`
}

//define a function for the default message handler
var mqttMessageHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
  fmt.Printf("TOPIC: %s\n", msg.Topic())

/*
/esp/water-c/stat
	{"act": "status", "time": 1534708109, "ap": "olans", "rssi": -67, "free": 38784, "vcc": 3.17, "uptime": 462600773}
/esp/init
	{"act":"connected", "ap": "olans", "name": "water-c", "mac": "FF:FF:FF:99:99:99", "ip": "192.168.1.29", "rssi": -71, "vcc": 3.08, "h": 0.9, "v": 0.9, "upd": 1534245504, "time": 1534245512, "inited": 4525, "free": 39616}
*/
  payload := []byte(msg.Payload())

  if !json.Valid(payload) {
  	fmt.Printf("MSG: %s\n", payload)
	return
  }

  var dat map[string]interface{}
  if err := json.Unmarshal(payload, &dat); err != nil {
    panic(err)
  }
  fmt.Println(dat)

  opcode := dat["act"].(string)
  switch opcode {
  case "connected":
      packet := mqttConnectedPacket{}
      json.Unmarshal(payload, &packet)
	  fmt.Printf("[%s]\n", packet.Name)
	  channel := "/esp/"+packet.Name+"/+"

	  if token := client.Subscribe(channel, 0, nil); token.Wait() && token.Error() != nil {
		  log.Error(token.Error())
	  } else {
		  fmt.Printf("Subscribed: %s\n", channel);
	  }
  case "status":
  	  packet := mqttStatusPacket{}
      json.Unmarshal(payload, &packet)
      fmt.Printf("Status[%d]: %s R:%d free:%d vcc:%.2f uptime:%d\n", packet.Time, packet.Ap, packet.Rssi, packet.Free, packet.Vcc, packet.Uptime);
  case "debug":
	  fmt.Println(opcode)
  }
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
