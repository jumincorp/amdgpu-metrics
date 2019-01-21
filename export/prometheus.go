package export

import (
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	collectors = make(map[string](*prometheus.GaugeVec))
)

type Prometheus struct {
	address string
	Exporter
}

func init() {
	collectors[clock] =
		prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "amdgpu_clock", Help: "GPU Clock Rate in MHz"}, []string{"gpu", "name"})

	collectors[power] =
		prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "amdgpu_power", Help: "GPU Power Consumption in Watts"}, []string{"gpu", "name"})

	collectors[temp] =
		prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "amdgpu_temp", Help: "GPU Temperature in Celcius"}, []string{"gpu"})

	collectors[load] =
		prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "amdgpu_load", Help: "GPU Load Percentage"}, []string{"gpu"})

	for _, c := range collectors {
		prometheus.MustRegister(c)
	}
}

func NewPrometheus(address string) *Prometheus {
	p := new(Prometheus)
	p.address = address
	return p
}

func (p *Prometheus) Setup() {
	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(p.address, nil))
}

func (p *Prometheus) Export(ctype string, gpu string, name string, value string) error {
	fValue, err := strconv.ParseFloat(value, 64)
	if err == nil {
		switch ctype {
		case clock, power:
			collectors[ctype].With(prometheus.Labels{"gpu": gpu, "name": name}).Set(fValue)
		default:
			collectors[ctype].With(prometheus.Labels{"gpu": gpu}).Set(fValue)
		}
	}
	return err
}
