package main
/* vim: set ts=2 sw=2 sts=2 et: */

import (
  "strings"
  "time"
  "fmt"
  "strconv"
  "encoding/json"
)

type espConnectedPacket struct {
    Act    string  `json:"act"`
    Name   string  `json:"name"`
    Ap     string  `json:"ap"`
    Mac    string  `json:"mac"`
    IP     string  `json:"ip"`
    H      string  `json:"h"`
    V      string  `json:"v"`
    Time   uint32  `json:"time"`
    UPD    uint32  `json:"upd"`
    Inited uint32  `json:"inited"`
    Free   uint32  `json:"free"`
    Vcc    float32 `json:"vcc"`
    Rssi   int16   `json:"rssi"`
}

/*
type espChannelsPacket struct {
    Act   string  `json:"act"`
    Chs   []string `json:"chs"`
}
*/
type espStatusPacket struct{
    Act    string  `json:"act"`
    Ap     string  `json:"ap"`
    Time   uint32  `json:"time"`
    Free   uint32  `json:"free"`
    Uptime uint32  `json:"uptime"`
    Vcc    float32 `json:"vcc"`
    Rssi   int16   `json:"rssi"`
}

var espKnownNodes = make(map[string]bool)

func espInitHandler(payload []byte) {
  packet := espConnectedPacket{}
  if err := json.Unmarshal(payload, &packet); err != nil {
    log.Error("INIT json.Unmarshal:", err)
  }

  verbosePrint("init "+packet.Name)
  n, ok := espKnownNodes[packet.Name]
  if (!n || !ok) {
    espNodeSubscribe(packet.Name)
  }
}

func espStatHandler(payload []byte, espName string) {
  packet := espStatusPacket{}
  if err := json.Unmarshal(payload, &packet); err != nil {
    log.Errorf("STAT [%s] json.Unmarshal: %v", espName, err)
  }

  log.Infof("(%s) status [%d]: %s R:%d free:%d vcc:%.2f uptime:%d",
    espName, packet.Time, packet.Ap, packet.Rssi, packet.Free, packet.Vcc, packet.Uptime);
}


func espMessageHandler(topic string, payload []byte) {
    messageTime := time.Now()
    verbosePrint("TOPIC: "+topic)
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
        log.Warnf("Unknown prefix in topic: %s", topic)
        return;
    }

    if espName == "init" {
      if json.Valid(payload) {
        espInitHandler(payload)
      } else {
        log.Warnf("INIT invalid json: %v", payload)
        return
      }

    } else if espTheme == "stat" {
      if json.Valid(payload) {
        espStatHandler(payload, espName)
      } else {
        log.Warnf("STAT invalid json: %v", payload)
        return
      }

    } else {
      payloadStr := string(payload[:])
      var espRoom = espName;
      if (len(espTag) > 0) {
          espRoom = espTag
      }

      var timestamp = messageTime.Unix()
      verbosePrint(fmt.Sprintf("[%d] %s %s %s", timestamp, espRoom, espTheme, payloadStr))

      switch espTheme {
        case "count", "move", "led", "temp", "humd", "pres":
          // int8
          if espTheme == "count" || espTheme == "move" || espTheme == "led" {
            value, err := strconv.ParseUint(payloadStr, 10, 64)
            if err != nil {
              // handle error
              log.Errorf("[%d] %s %s %s convert to int: %s", timestamp, espRoom, espTheme, payloadStr, err)
              return
            }

            if value > 1000000000 {
              timestamp = int64(value)
            }

            dbAddEvent(espRoom, espTheme, value, timestamp)
          // float
          } else if espTheme == "temp" || espTheme == "humd" || espTheme == "pres" {
            value, err := strconv.ParseFloat(payloadStr, 64)
            if err != nil {
              // handle error
              log.Errorf("[%d] %s %s %s convert to float: %s", timestamp, espRoom, espTheme, payloadStr, err)
              return
            }

            dbAddMetric(espRoom, espTheme, value, timestamp)
          // string
          } else {
            //dbAdd(espRoom, espTheme, payloadStr, timestamp)
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
          log.Warnf("Unknown topic Theme (%s) [%u] %s %s %s %u,"+
            " we ignore it (but devs must fix this by adding esp handler for it)",
            espTheme, timestamp, espRoom, espTheme, payloadStr)
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

