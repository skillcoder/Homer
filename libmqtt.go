package main

/* vim: set ts=2 sw=2 sts=2 ff=unix ft=go noet: */

import (
	"fmt"
	"time"

	"github.com/goiiot/libmqtt"
)

var mqttClient libmqtt.Client

func mqttSubscribe(name string) {
	log.Debugf("Subscribing: %s", name)
	mqttClient.Subscribe(&libmqtt.Topic{Name: name})
}

func pubHandler(topic string, err error) {
	if err != nil {
		log.Error("PubErr:", err)
	} else {
		log.Infof("Pub: %s", topic)
	}
}

func subHandler(topics []*libmqtt.Topic, err error) {
	if err != nil {
		log.Error("SubErr:", err)
		// TODO: We musr recover this first!!!
	} else {
		for _, topic := range topics {
			log.Infof("Subscribed: %s", topic)
		}
	}
}

func unSubHandler(topics []string, err error) {
	if err != nil {
		log.Error("UnSubErr:", err)
	} else {
		for _, topic := range topics {
			log.Infof("UnSubscribed: %s", topic)
		}
	}
}

func netHandler(server string, err error) {
	if err != nil {
		log.Error("NetErr:", server, err)
	} else {
		log.Infof("ServerConnected: %s", server)
	}
}

func persistHandler(err error) {
	if err != nil {
		log.Error("PersistErr:", err)
	} else {
		log.Infof("PersistHandler")
	}
}

func mqttConnect(host string, port uint16, name, user, pass string) {
	mqtturi := fmt.Sprintf("%s:%d", host, port)
	var err error
	mqttClient, err = libmqtt.NewClient(
		libmqtt.WithServer(mqtturi),
		libmqtt.WithBuf(512, 512), // send, recv
		libmqtt.WithClientID(name),
		libmqtt.WithDialTimeout(2), // Connection timeout
		libmqtt.WithIdentity(user, pass),
		libmqtt.WithKeepalive(30, 1),
		libmqtt.WithLog(libmqtt.Info), // Silent/Verbose/Debug/Info/Warning/Error
		libmqtt.WithRouter(libmqtt.NewRegexRouter()),
		//libmqtt.WithPersist
		libmqtt.WithBackoffStrategy(1*time.Second, 30*time.Second, 1.5),
		libmqtt.WithAutoReconnect(true),
	)
	if err != nil {
		// handle client creation error
		log.Panic("mqttClient creation ERR:", err)
	}

	mqttClient.HandlePub(pubHandler)
	mqttClient.HandleSub(subHandler)
	mqttClient.HandleUnSub(unSubHandler)
	// register handler for net error (optional, but recommended)
	mqttClient.HandleNet(netHandler)
	// register handler for persist error (optional, but recommended)
	mqttClient.HandlePersist(persistHandler)
	mqttClient.Handle(".*",
		func(topic string, qos libmqtt.QosLevel, msg []byte) {
			espMessageHandler(topic, msg)
		})
	// connect to server
	mqttClient.Connect(func(server string, code byte, err error) {
		if err != nil {
			log.Error("mqttConnect ERR:", err)
			return
		}

		if code != libmqtt.CodeSuccess {
			// server rejected or in error
			log.Panicf("mqttConnect server CODE: %d", code)
		}

		espInit()
	})
}
