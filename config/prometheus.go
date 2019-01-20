package config

import "github.com/spf13/viper"

const (
	cfgPrometheusAddress = "Prometheus.Address"
)

// Prometheus represents the Prometheus section of the configuration
type Prometheus struct {
}

func newPrometheus() *Prometheus {
	prometheus := new(Prometheus)
	viper.SetDefault(cfgPrometheusAddress, ":40011")

	return prometheus
}

func (prometheus *Prometheus) Address() string {
	return viper.Get(cfgPrometheusAddress).(string)
}
