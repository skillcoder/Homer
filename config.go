/* vim: set ts=2 sw=2 sts=2 et: */
package main

import (
  "io/ioutil"
  "os"
  "strconv"
  "gopkg.in/yaml.v2"
)

type config_t struct {
  ConfigDir string `yaml:"configdir"`
  DataDir string `yaml:"datadir"`
  Mode    string `yaml:"mode"`
  Listen  string `yaml:"listen"`
  Mqtt struct {
    Host string `yaml:"host"`
    Port uint16 `yaml:"port"`
    User string `yaml:"user"`
    Pass string `yaml:"pass"`
    Name string `yaml:"name"`
  }
  ClickHouse struct {
    Host string `yaml:"host"`
    Port uint16 `yaml:"port"`
    User string `yaml:"user"`
    Pass string `yaml:"pass"`
    Name string `yaml:"name"`
  }
}

var config config_t = config_t{}

func config_load() {
  config_set_from_env("HOMER_CONFIG_DIR",  "/etc/homer/", false)
  configfile := config.ConfigDir+"config.yml"
  if _, err := os.Stat(configfile); os.IsNotExist(err) {
    configfile = "config.yml"
  }

  log.Debugf("Read config: %s", configfile)
  data := read_file(configfile)

  err := yaml.Unmarshal([]byte(data), &config)
  if err != nil {
    log.Fatalf("error: %v", err)
    os.Exit(2)
  }
/*
  var allenvlist = [...]string {
    "HOMER_CONFIG_DIR"
    "HOMER_DATA_DIR",
    "HOMER_MODE",
    "HOMER_LISTEN",
    "HOMER_MQTT_HOST",
    "HOMER_MQTT_PORT",
    "HOMER_MQTT_USER",
    "HOMER_MQTT_PASS",
    "HOMER_MQTT_NAME",
    "HOMER_CLICKHOUSE_HOST",
    "HOMER_CLICKHOUSE_PORT",
    "HOMER_CLICKHOUSE_USER",
    "HOMER_CLICKHOUSE_PASS",
    "HOMER_CLICKHOUSE_NAME",
  }
*/

  //log.Debugf("CONFIG: %#v", config)
  config_set_from_env("HOMER_CONFIG_DIR",  "/etc/homer/", true)
  config_set_from_env("HOMER_DATA_DIR",  "/var/lib/homer/", true)
  config_set_from_env("HOMER_MODE",      "production", true)
  config_set_from_env("HOMER_LISTEN",    ":18266", true)
  config_set_from_env("HOMER_MQTT_HOST", "127.0.0.1", true)
  config_set_from_env("HOMER_MQTT_PORT", 1883, true)
  config_set_from_env("HOMER_MQTT_USER", "", true)
  config_set_from_env("HOMER_MQTT_PASS", "", false)
  config_set_from_env("HOMER_MQTT_NAME", "go-homer-server", true)
  config_set_from_env("HOMER_CLICKHOUSE_HOST", "127.0.0.1", true)
  config_set_from_env("HOMER_CLICKHOUSE_PORT", 9000, true)
  config_set_from_env("HOMER_CLICKHOUSE_USER", "homer", true)
  config_set_from_env("HOMER_CLICKHOUSE_PASS", "", false)
  config_set_from_env("HOMER_CLICKHOUSE_NAME", "homer", true)
}

func config_set_from_env(envname string, default_value interface{}, is_log bool) {
  switch envname {
    case "HOMER_CONFIG_DIR":
      if len(os.Getenv(envname)) > 0 { config.ConfigDir = os.Getenv(envname) }
      if len(config.ConfigDir) == 0 { config.ConfigDir = default_value.(string) }
      if is_log {log.Debugf("%s = %s", envname, config.ConfigDir)}
    case "HOMER_DATA_DIR":
      if len(os.Getenv(envname)) > 0 { config.DataDir = os.Getenv(envname) }
      if len(config.DataDir) == 0 { config.DataDir = default_value.(string) }
      if is_log {log.Debugf("%s = %s", envname, config.DataDir)}
    case "HOMER_MODE":
      if len(os.Getenv(envname)) > 0 { config.Mode = os.Getenv(envname) }
      if len(config.Mode) == 0 { config.Mode = default_value.(string) }
      if is_log {log.Debugf("%s = %s", envname, config.Mode)}
    case "HOMER_LISTEN":
      if len(os.Getenv(envname)) > 0 { config.Listen = os.Getenv(envname) }
      if len(config.Listen) == 0 { config.Listen = default_value.(string) }
      if is_log {log.Debugf("%s = %s", envname, config.Listen)}
    case "HOMER_MQTT_HOST":
      if len(os.Getenv(envname)) > 0 { config.Mqtt.Host = os.Getenv(envname) }
      if len(config.Mqtt.Host) == 0 { config.Mqtt.Host = default_value.(string) }
      if is_log {log.Debugf("%s = %s", envname, config.Mqtt.Host)}
    case "HOMER_MQTT_PORT":
      if len(os.Getenv(envname)) > 0 {
        if val, err := strconv.ParseUint(os.Getenv(envname), 10, 16); err == nil {
          config.Mqtt.Port = uint16(val)
        } else {
          log.Panicf("Cant parse %s as uint16: %v", envname, err)
        }
      }
      if config.Mqtt.Port == 0 { config.Mqtt.Port = default_value.(uint16) }
      if is_log {log.Debugf("%s = %d", envname, config.Mqtt.Port)}
    case "HOMER_MQTT_USER":
      if len(os.Getenv(envname)) > 0 { config.Mqtt.User = os.Getenv(envname) }
      if len(config.Mqtt.User) == 0 { config.Mqtt.User = default_value.(string) }
      if is_log {log.Debugf("%s = %s", envname, config.Mqtt.User)}
    case "HOMER_MQTT_PASS":
      if len(os.Getenv(envname)) > 0 { config.Mqtt.Pass = os.Getenv(envname) }
      if len(config.Mqtt.Pass) == 0 { config.Mqtt.Pass = default_value.(string) }
      if is_log {log.Debugf("%s = %s", envname, config.Mqtt.Pass)}
    case "HOMER_MQTT_NAME":
      if len(os.Getenv(envname)) > 0 { config.Mqtt.Name = os.Getenv(envname) }
      if len(config.Mqtt.Name) == 0 { config.Mqtt.Name = default_value.(string) }
      if is_log {log.Debugf("%s = %s", envname, config.Mqtt.Name)}
    case "HOMER_CLICKHOUSE_HOST":
      if len(os.Getenv(envname)) > 0 { config.ClickHouse.Host = os.Getenv(envname) }
      if len(config.ClickHouse.Host) == 0 { config.ClickHouse.Host = default_value.(string) }
      if is_log {log.Debugf("%s = %s", envname, config.ClickHouse.Host)}
    case "HOMER_CLICKHOUSE_PORT":
      if len(os.Getenv(envname)) > 0 {
        if val, err := strconv.ParseUint(os.Getenv(envname), 10, 16); err == nil {
          config.ClickHouse.Port = uint16(val)
        } else {
          log.Panicf("Cant parse %s as uint16: %v", envname, err)
        }
      }
      if config.ClickHouse.Port == 0 { config.ClickHouse.Port = default_value.(uint16) }
      if is_log {log.Debugf("%s = %d", envname, config.ClickHouse.Port)}
    case "HOMER_CLICKHOUSE_USER":
      if len(os.Getenv(envname)) > 0 { config.ClickHouse.User = os.Getenv(envname) }
      if len(config.ClickHouse.User) == 0 { config.ClickHouse.User = default_value.(string) }
      if is_log {log.Debugf("%s = %s", envname, config.ClickHouse.User)}
    case "HOMER_CLICKHOUSE_PASS":
      if len(os.Getenv(envname)) > 0 { config.ClickHouse.Pass = os.Getenv(envname) }
      if len(config.ClickHouse.Pass) == 0 { config.ClickHouse.Pass = default_value.(string) }
      if is_log {log.Debugf("%s = %s", envname, config.ClickHouse.Pass)}
    case "HOMER_CLICKHOUSE_NAME":
      if len(os.Getenv(envname)) > 0 { config.ClickHouse.Name = os.Getenv(envname) }
      if len(config.ClickHouse.Name) == 0 { config.ClickHouse.Name = default_value.(string) }
      if is_log {log.Debugf("%s = %s", envname, config.ClickHouse.Name)}
      /*
    case "":
      if len(os.Getenv(envname)) > 0 { config. = os.Getenv(envname) }
      if len(config.) == 0 { config. = default_value.(string) }
      */
    default:
      log.Panic("Unknown envname: ", envname)
    }
}

func read_file(filename string) (b []byte) {
  b, err := ioutil.ReadFile(filename)
  if err != nil {
    log.Panic(err)
    panic(err);
  }

  return
}
