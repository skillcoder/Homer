/* vim: set ts=2 sw=2 sts=2 et: */
package main

import (
  "fmt"
  "github.com/goiiot/libmqtt"
)

var mqttClient libmqtt.Client

func mqttSubscribe(Name string) {
  mqttClient.Subscribe(&libmqtt.Topic{Name: Name})
}

func PubHandler (topic string, err error) {
  if err != nil {
    log.Error("PubErr:", err)
  } else {
    log.Infof("Pub: %s", topic);
  }
}

func SubHandler (topics []*libmqtt.Topic, err error) {
  if err != nil {
    log.Error("SubErr:", err)
    // TODO: We musr recover this first!!!
  } else {
    for _, topic := range(topics) {
      mqttClient.Handle(topic.Name, func(topic string, qos libmqtt.QosLevel, msg []byte) {
        espMessageHandler(topic, msg)
      })
      log.Infof("Subscribed: %s", topic);
    }
  }
}

func UnSubHandler (topics []string, err error) {
  if err != nil {
    log.Error("UnSubErr:", err)
  } else {
    for _, topic := range(topics) {
      log.Infof("UnSubscribed: %s", topic);
    }
  }
}

func NetHandler(server string, err error) {
  if err != nil {
    log.Error("NetErr:", server, err)
  } else {
    log.Infof("ServerConnected: %s", server);
  }
}

func PersistHandler(err error) {
  if err != nil {
    log.Error("PersistErr:", err)
  } else {
    log.Infof("PersistHandler");
  }
}

func mqttConnect(host string, port uint16, name string) {
  mqtturi := fmt.Sprintf("%s:%d", host, port);
  mqttClient, err := libmqtt.NewClient(
    libmqtt.WithServer(mqtturi),
    libmqtt.WithBuf(512, 512), // send, recv
    libmqtt.WithClientID(name),
    libmqtt.WithDialTimeout(2), // Connection timeout
    //libmqtt.WithIdentity(username, password),
//    libmqtt.WithKeepalive(30, 1),
    libmqtt.WithLog(libmqtt.Verbose), // Silent/Verbose/Debug/Info/Warning/Error
    //libmqtt.WithPersist
  )
  if err != nil {
    // handle client creation error
    log.Panic("mqttClient creation ERR:", err)
  }

  mqttClient.HandlePub(PubHandler)
  mqttClient.HandleSub(SubHandler)
  mqttClient.HandleUnSub(UnSubHandler)
  // register handler for net error (optional, but recommended)
  mqttClient.HandleNet(NetHandler)
  // register handler for persist error (optional, but recommended)
  mqttClient.HandlePersist(PersistHandler)
  // define your topic handlers like a golang http server
//  mqttClient.Handle("foo", func(topic string, qos libmqtt.QosLevel, msg []byte) {
    // handle the topic message
//  })
  // connect to server
  mqttClient.Connect(func(server string, code byte, err error) {
    if err != nil {
        log.Panic("mqttConnect ERR:", err)
    }

    if code != libmqtt.CtrlConn {
        // server rejected or in error
        log.Panicf("mqttConnect server CODE: %d", code)
    }

    espInit()
  })
}
