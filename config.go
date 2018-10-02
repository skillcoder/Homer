package main

/* vim: set ts=2 sw=2 sts=2 ff=unix ft=go noet: */

import (
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

type configT struct {
	ConfigFile      string `mapstructure:"config_file"`
	ConfigDir       string `mapstructure:"config_dir"`
	DataDir         string `mapstructure:"data_dir"`
	Mode            string `mapstructure:"mode"`
	Listen          string `mapstructure:"listen"`
	AggregatePeriod uint32 `mapstructure:"aggregate_period"`
	Verbose         bool   `mapstructure:"verbose"`
	Mqtt            struct {
		Host string `mapstructure:"host"`
		Port uint16 `mapstructure:"port"`
		User string `mapstructure:"user"`
		Pass string `mapstructure:"pass"`
		Name string `mapstructure:"name"`
	}
	ClickHouse struct {
		Host string `mapstructure:"host"`
		Port uint16 `mapstructure:"port"`
		User string `mapstructure:"user"`
		Pass string `mapstructure:"pass"`
		Name string `mapstructure:"name"`
	}
	Counters struct {
		WaterC []string `mapstructure:"water-c"`
		WaterH []string `mapstructure:"water-h"`
	}
}

var config = configT{}
var configCounters = make(map[string]uint8)

// TODO: switch to github.com/spf13/viper
func configLoad() {
	// use config file if set from the flags
	if config.ConfigFile != "" {
		log.Debug("Using config file ", config.ConfigFile)
		viper.SetConfigFile(config.ConfigFile)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	var DefaultConfigDir = "/etc/homer"
	var DefaultDataDir = "/var/lib/homer/"
	if runtime.GOOS == "freebsd" {
		DefaultConfigDir = "/usr/local/etc/homer"
		DefaultDataDir = "/var/db/homer/"
	}

	log.Debugf("Add %s config path %s", runtime.GOOS, DefaultConfigDir)
	viper.AddConfigPath(DefaultConfigDir)
	viper.AddConfigPath(".")

	viper.SetDefault("ConfigDir", DefaultConfigDir)
	viper.SetDefault("DataDir", DefaultDataDir)
	viper.SetDefault("Mode", "production")
	viper.SetDefault("Listen", ":18266")
	viper.SetDefault("AggregatePeriod", 5000)
	viper.SetDefault("Mqtt.Host", "127.0.0.1")
	viper.SetDefault("Mqtt.Port", 1883)
	viper.SetDefault("Mqtt.User", "")
	viper.SetDefault("Mqtt.Pass", "")
	viper.SetDefault("Mqtt.Name", "go-homer-server")
	viper.SetDefault("ClickHouse.Host", "127.0.0.1")
	viper.SetDefault("ClickHouse.Port", 9000)
	viper.SetDefault("ClickHouse.User", "homer")
	viper.SetDefault("ClickHouse.Pass", "")
	viper.SetDefault("ClickHouse.Name", "homer")

	err := viper.ReadInConfig()
	switch err.(type) {
	case viper.UnsupportedConfigError:
		log.Info("No config file, using defaults")
	default:
		check(err)
	}

	config.ConfigFile = viper.ConfigFileUsed()
	// TODO: relative to the config/binary -> https://github.com/davidpelaez/gh-keys/blob/master/gh-keys/config.go

	viper.SetEnvPrefix("HOMER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	//BindEnv("")
	viper.AutomaticEnv()

	err = viper.Unmarshal(&config)
	check(err)

	configPrintSummary()

	//make optimize varibles
	for _, v := range config.Counters.WaterC {
		configCounters[v] = 1 // water cold
	}

	for _, v := range config.Counters.WaterH {
		configCounters[v] = 2 // water hot
	}
}

func logConfigItem(key string, rawValue interface{}) {
	log.Debugf("%19s: %v", key, rawValue)
}

func configPrintSummary() {
	logConfigItem("ConfigFile", config.ConfigFile)
	logConfigItem("ConfigDir", config.ConfigDir)
	logConfigItem("DataDir", config.DataDir)
	logConfigItem("Mode", config.Mode)
	logConfigItem("Listen", config.Listen)
	logConfigItem("AggregatePeriod", config.AggregatePeriod)
	logConfigItem("Verbose", config.Verbose)
	logConfigItem("Mqtt.Host", config.Mqtt.Host)
	logConfigItem("Mqtt.Port", config.Mqtt.Port)
	logConfigItem("Mqtt.User", config.Mqtt.User)
	logConfigItem("Mqtt.Name", config.Mqtt.Name)
	logConfigItem("ClickHouse.Host", config.ClickHouse.Host)
	logConfigItem("ClickHouse.Port", config.ClickHouse.Port)
	logConfigItem("ClickHouse.User", config.ClickHouse.User)
	logConfigItem("ClickHouse.Name", config.ClickHouse.Name)
	logConfigItem("Counters.WaterC", config.Counters.WaterC)
	logConfigItem("Counters.WaterH", config.Counters.WaterH)
}
