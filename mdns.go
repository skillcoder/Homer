package main
/* vim: set ts=2 sw=2 sts=2 noexpandtab: */

import (
  "strings"

  "github.com/micro/mdns"
)

func mdnsInit() {
  // Make a channel for results and start listening
  entriesCh := make(chan *mdns.ServiceEntry, 8)
  go func() {
    for entry := range entriesCh {
      //fmt.Printf("Got new entry: %v\n", entry)
      log.Infof("New node detected: (%s) [%s]", entry.Host, entry.AddrV4)
      s := strings.SplitN(entry.Host, ".", 2)
      espNodeSubscribe(s[0])
    }

    log.Info("Node discovery thread ended")
  }()

  // Start the lookup
  err := mdns.Lookup("_homer._tcp", entriesCh)
  if err != nil {
    log.Error(err)
  }

  close(entriesCh)
}
