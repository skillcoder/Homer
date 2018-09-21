/* vim: set ts=2 sw=2 sts=2 et: */
package main

import (
	"encoding/json"
	"strings"
  "time"
  "strconv"
)

type espConnectedPacket struct {
    Act    string  `json:"act"`
    Time   uint32  `json:"time"`
    Name   string  `json:"name"`
    Ap     string  `json:"ap"`
    Mac    string  `json:"mac"`
    Ip     string  `json:"ip"`
    H      string  `json:"h"`
    V      string  `json:"v"`
    UPD    uint32  `json:"upd"`
    Inited uint32  `json:"inited"`

    Rssi   int16   `json:"rssi"`
    Vcc    float32 `json:"vcc"`
    Free   uint32  `json:"free"`
}

/*
type espChannelsPacket struct {
    Act   string  `json:"act"`
    Chs   []string `json:"chs"`
}
*/
type espStatusPacket struct{
    Act    string  `json:"act"`
    Time   uint32  `json:"time"`
    Ap     string  `json:"ap"`
    Rssi   int16   `json:"rssi"`
    Vcc    float32 `json:"vcc"`
    Free   uint32  `json:"free"`
    Uptime uint32  `json:"uptime"`
}

var espKnownNodes map[string]bool = make(map[string]bool)

func espMessageHandler(topic string, payload []byte) {
    messageTime := time.Now()
    log.Debugf("TOPIC: %s", topic)
    s := strings.Split(topic, "/")
    var espName string
    var espTheme string
    var espTag string
    if (s[1] == "esp") {
        espName = s[2]
        if (len(s) > 3) {
            espTheme = s[3]
        }
        if (len(s) > 4) {
            espTag = s[4]
        }
    } else {
        return;
    }

    if (espName == "init" || espTheme == "stat") {
        if !json.Valid(payload) {
            log.Warnf("MSG: %s", payload)
            return
        }

        var dat map[string]interface{}
        if err := json.Unmarshal(payload, &dat); err != nil {
            panic(err)
        }
        //fmt.Println(dat)

        opcode := dat["act"].(string)
        switch opcode {
        case "connected":
            packet := espConnectedPacket{}
            json.Unmarshal(payload, &packet)
            log.Debugf("[%s]", packet.Name)
            n, ok := espKnownNodes[packet.Name]
            //fmt.Printf("[%v] [%v]\n", n, ok)
            if (!n || !ok) {
                espNodeSubscribe(packet.Name)
            }
        case "status":
            packet := espStatusPacket{}
            json.Unmarshal(payload, &packet)
            log.Infof("(%s) status [%d]: %s R:%d free:%d vcc:%.2f uptime:%d", espName, packet.Time, packet.Ap, packet.Rssi, packet.Free, packet.Vcc, packet.Uptime);
            /*
        case "channels":
            packet := espChannelsPacket{}
            json.Unmarshal(payload, &packet)
            */
        default:
            log.Warnf("Unimplemented opcode: %s", payload)
        }
    } else {
      payload_str := string(payload[:])
      var espRoom string = espName;
      if (len(espTag) > 0) {
          espRoom = espTag
      }

      var timestamp int64 = messageTime.Unix()
      log.Debugf("[%d] %s %s %s", timestamp, espRoom, espTheme, payload_str)

      switch espTheme {
        case "count", "move", "led", "temp", "humd", "pres":
          // int8
          if espTheme == "count" || espTheme == "move" || espTheme == "led" {
            value, err := strconv.ParseUint(payload_str, 10, 64)
            if err != nil {
              // handle error
              log.Errorf("[%d] %s %s %s convert to int: %s", timestamp, espRoom, espTheme, payload_str, err)
              return
            }

            if value > 1000000000 {
              timestamp = int64(value)
            }

            dbAddEvent(espRoom, espTheme, value, timestamp)
          // float
          } else if espTheme == "temp" || espTheme == "humd" || espTheme == "pres" {
            value, err := strconv.ParseFloat(payload_str, 64)
            if err != nil {
              // handle error
              log.Errorf("[%d] %s %s %s convert to float: %s", timestamp, espRoom, espTheme, payload_str, err)
              return
            }

            dbAddMetric(espRoom, espTheme, value, timestamp)
          // string
          } else {
            //dbAdd(espRoom, espTheme, payload_str, timestamp)
            log.Errorf("Unknown Theme type: %s", espTheme)
          }
/*
        switch espTheme {
        // Counter (unix_timestamp of click)
        case "count":
        // Movement detector (unix_timestamp/0) 0 = end detection
        case "move":
        // Switch
        case "led":
        // Temperature
        case "temp":
        // Humidity
        case "humd":
        // Pressure
        case "pres":
        }
*/
        case "debug":
          //log.Debugf("Debug")
        default:
          log.Warnf("Unknown topic Theme (%s) [%u] %s %s %s %u, we ignore it (but devs must fix this by adding esp handler for it)", espTheme, timestamp, espRoom, espTheme, payload_str)
      }
   }
}

func espNodeSubscribe(Name string) {
    espKnownNodes[Name] = true
    channel := "/esp/"+Name+"/#"
    mqttSubscribe(channel)
}

func espInit() {
  mqttSubscribe("/esp/init")
  go mdnsInit()
}

