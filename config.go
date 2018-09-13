/* vim: set ts=2 sw=2 sts=2 et: */
package main

import (
	"io/ioutil"
  "os"
  "gopkg.in/yaml.v2"
)

type config_t struct {
  Mqtt struct {
    Host string `yaml:"host"`
    Port int16  `yaml:"port"`
    User string `yaml:"user"`
    Pass string `yaml:"pass"`
  }
  ClickHouse struct {
    Host string `yaml:"host"`
    Port int16  `yaml:"port"`
    User string `yaml:"user"`
    Pass string `yaml:"pass"`
    Name string `yaml:"name"`
  }
}

var config config_t = config_t{}

func config_load() {
  data := read_file("config.yml")

  err := yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		log.Fatalf("error: %v", err)
    os.Exit(2)
	}

  //log.Debugf("CONFIG: %#v", config)
}

func read_file(filename string) (b []byte) {
  b, err := ioutil.ReadFile(filename)
  if err != nil {
    log.Panic(err)
    panic(err);
  }

  return
}
