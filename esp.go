package main

/* vim: set ts=2 sw=2 sts=2 ff=unix ft=go noet: */

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
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
type espStatusPacket struct {
	Act    string  `json:"act"`
	Ap     string  `json:"ap"`
	Time   uint32  `json:"time"`
	Free   uint32  `json:"free"`
	Uptime uint32  `json:"uptime"`
	Vcc    float32 `json:"vcc"`
	Rssi   int16   `json:"rssi"`
}

var espKnownNodes = make(map[string]bool)

func espHandleInit(payload []byte) {
	if !json.Valid(payload) {
		errStr := fmt.Sprintf("INIT invalid json: %v", payload)
		log.Warn(errStr)
		return
	}

	packet := espConnectedPacket{}
	if err := json.Unmarshal(payload, &packet); err != nil {
		log.Error("INIT json.Unmarshal:", err)
		return
	}

	verbosePrint("init " + packet.Name)
	n, ok := espKnownNodes[packet.Name]
	if !n || !ok {
		espNodeSubscribe(packet.Name)
	}
}

func espHandleStat(payload []byte, espName string) {
	if !json.Valid(payload) {
		errStr := fmt.Sprintf("STAT invalid json: %v", payload)
		log.Warn(errStr)
		return
	}

	packet := espStatusPacket{}
	if err := json.Unmarshal(payload, &packet); err != nil {
		log.Errorf("STAT [%s] json.Unmarshal: %v", espName, err)
		return
	}

	log.Infof("(%s) status [%d]: %s R:%d free:%d vcc:%.2f uptime:%d",
		espName, packet.Time, packet.Ap, packet.Rssi, packet.Free, packet.Vcc, packet.Uptime)
}

func espHandleEvent(espRoom, espTheme, payloadStr string, timestamp int64) {
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
}

func espHandleMetric(espRoom, espTheme, payloadStr string, timestamp int64) {
	value, err := strconv.ParseFloat(payloadStr, 64)
	if err != nil {
		// handle error
		log.Errorf("[%d] %s %s %s convert to float: %s", timestamp, espRoom, espTheme, payloadStr, err)
		return
	}

	dbAddMetric(espRoom, espTheme, value, timestamp)
}

func parseTopic(topic string) (espName, espTheme, espTag string, err error) {
	s := strings.Split(topic, "/")
	if s[1] == "esp" {
		espName = s[2]
		if len(s) > 3 {
			espTheme = s[3]
		}

		if len(s) > 4 {
			espTag = s[4]
		}
	} else {
		return "", "", "", errors.New("Unknown prefix in topic: " + topic)
	}

	return espName, espTheme, espTag, nil
}

func espMessageHandler(topic string, payload []byte) {
	messageTime := time.Now()
	verbosePrint("TOPIC: " + topic)
	espName, espTheme, espTag, err := parseTopic(topic)
	if err != nil {
		log.Warn(err)
		return
	}

	if espName == "init" {
		espHandleInit(payload)

	} else if espTheme == "stat" {
		espHandleStat(payload, espName)

	} else {
		payloadStr := string(payload[:])
		var espRoom = espName
		if len(espTag) > 0 {
			espRoom = espTag
		}

		var timestamp = messageTime.Unix()
		verbosePrint(fmt.Sprintf("[%d] %s %s %s", timestamp, espRoom, espTheme, payloadStr))

		switch espTheme {
		// int
		case "count", "move", "led":
			espHandleEvent(espRoom, espTheme, payloadStr, timestamp)
		// float
		case "temp", "humd", "pres":
			espHandleMetric(espRoom, espTheme, payloadStr, timestamp)

		case "debug":
			log.Debugf("Debug")

		default:
			log.Warnf("Unknown topic Theme (%s) [%u] %s %s %s %u,"+
				" we ignore it (but devs must fix this by adding esp handler for it)",
				espTheme, timestamp, espRoom, espTheme, payloadStr)
		}
	}
}

func espNodeSubscribe(Name string) {
	espKnownNodes[Name] = true
	channel := "/esp/" + Name + "/#"
	mqttSubscribe(channel)
}

func espInit() {
	mqttSubscribe("/esp/init")
	go mdnsInit()
}
