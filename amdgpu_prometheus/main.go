package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"../config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	programName = "amdgpu_prometheus"

	clock = "clock"
	power = "power"
	temp  = "temp"
	load  = "load"
)

type pmInfoFile struct {
	gpu  int
	path string
}

//
// Find all amd_gpu_pm_info files
//
func getpmInfoFiles(list *[]pmInfoFile) error {
	err := filepath.Walk("/sys/kernel/debug/dri", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "amdgpu_pm_info" {
			dir, _ := filepath.Split(path)
			dirList := strings.Split(strings.Trim(dir, "/"), "/")
			gpuString := dirList[len(dirList)-1]
			gpu, _ := strconv.Atoi(gpuString)

			*list = append(*list, pmInfoFile{gpu: gpu, path: path})
		}
		return nil
	})
	return err
}

func createCollectors(clocks map[string]string, collectors *map[string](*prometheus.GaugeVec)) {
	(*collectors)[clock] =
		prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "amdgpu_clock", Help: "GPU Clock Rate in MHz"}, []string{"gpu", "name"})

	(*collectors)[power] =
		prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "amdgpu_power", Help: "GPU Power Consumption in Watts"}, []string{"gpu", "name"})

	(*collectors)[temp] =
		prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "amdgpu_temp", Help: "GPU Temperature in Celcius"}, []string{"gpu"})

	(*collectors)[load] =
		prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "amdgpu_load", Help: "GPU Load Percentage"}, []string{"gpu"})

	for _, c := range *collectors {
		prometheus.MustRegister(c)
	}
}

func mapRegexp(text string, expression string) map[string]string {
	var res = make(map[string]string)

	r := regexp.MustCompile(expression)

	subexpNames := r.SubexpNames()

	for _, submatchList := range r.FindAllStringSubmatch(text, -1) {
		var m = make(map[string]string)
		for i, submatch := range submatchList[1:] {
			m[subexpNames[i+1]] = submatch
		}
		res[m["name"]] = m["val"]
	}
	return res
}

func main() {

	var (
		expressions = map[string]string{
			clock: `(?P<val>[0-9]+(?:\.[0-9]+)?) MHz \((?P<name>(?:[A-Za-z0-9\ ]+))\)`,
			power: `(?P<val>[0-9]+(?:\.[0-9]+)?) W \((?P<name>(?:[A-Za-z0-9\ ]+))\)`,
			temp:  `(?P<name>GPU Temperature): (?P<val>[0-9]+([0-9]+)?) C`,
			load:  `(?P<name>GPU Load): (?P<val>[0-9]+([0-9]+)?) %`,
		}

		pmInfoFileList []pmInfoFile
		cfg            *config.Config = config.NewConfig(programName)
		collectors                    = make(map[string](*prometheus.GaugeVec))
	)

	err := getpmInfoFiles(&pmInfoFileList)
	if err != nil {
		log.Fatal("Cannot read amdgpu_pm_info files\n")
	}

	go func() {
		for {
			for _, info := range pmInfoFileList {
				log.Printf("gpu %d, path %s", info.gpu, info.path)
				gpu := strconv.Itoa(info.gpu)

				bytes, err := ioutil.ReadFile(info.path)
				if err == nil {
					text := string(bytes)

					fmt.Println(text)

					if len(collectors) == 0 {
						createCollectors(mapRegexp(text, expressions[clock]), &collectors)
					}

					for _, ctype := range [2]string{clock, power} {
						for name, value := range mapRegexp(text, expressions[ctype]) {
							fValue, _ := strconv.ParseFloat(value, 64)
							collectors[ctype].With(prometheus.Labels{"gpu": gpu, "name": name}).Set(fValue)
						}
					}

					for _, ctype := range [2]string{temp, load} {
						for name, value := range mapRegexp(text, expressions[ctype]) {
							log.Printf("name %v value %v", name, value)
							fValue, _ := strconv.ParseFloat(value, 64)
							collectors[ctype].With(prometheus.Labels{"gpu": gpu}).Set(fValue)
						}
					}

				} else {
					log.Printf("Error reading file %v: %v\n", info.path, err)
				}
			}

			time.Sleep(time.Second * cfg.QueryDelay())
		}
	}()

	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(cfg.Prometheus.Address(), nil))
}
