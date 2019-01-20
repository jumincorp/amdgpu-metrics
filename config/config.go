package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

const (
	cfgQueryDelay = "QueryDelay"
)

// Config represents the configuration file. You should use NewConfig to create one.
type Config struct {
	Prometheus *Prometheus
	queryDelay time.Duration
}

// NewConfig creates an instance of the configuration
func NewConfig(name string) *Config {
	cfg := new(Config)

	viper.SetConfigName(name)
	viper.AddConfigPath(fmt.Sprintf("/etc/%v", name))

	cfg.Prometheus = newPrometheus()

	viper.SetDefault(cfgQueryDelay, 15)

	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		fmt.Printf("error reading config file: %s\n", err)
	}
	return cfg
}

// QueryDelay returns the time we wait to interrogate the miner again
func (cfg *Config) QueryDelay() time.Duration {
	return time.Duration(viper.Get(cfgQueryDelay).(int))
}
