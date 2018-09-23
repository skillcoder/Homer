package main

/* vim: set ts=2 sw=2 sts=2 ff=unix noexpandtab: */

import (
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

type configT struct {
	ConfigFile      string `yaml:"config_file"`
	ConfigDir       string `yaml:"config_dir"`
	DataDir         string `yaml:"data_dir"`
	Mode            string `yaml:"mode"`
	Listen          string `yaml:"listen"`
	AggregatePeriod uint32 `yaml:"aggregate_period"`
	Verbose         bool   `yaml:"verbose"`
	Mqtt            struct {
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

var config = configT{}

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
}
