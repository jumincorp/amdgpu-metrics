package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"../config"
)

const (
	programName = "amdgpu_prometheus"

	clock = "clock"
	power = "power"
	temp  = "temp"
	load  = "load"
)

var (
	minerGpuHashRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "miner_gpu_hashrate",
			Help: "Current hash rate of a GPU.",
		},
		[]string{"miner", "gpu", "symbol"},
	)

	cfg *config.Config
)

func init() {
	cfg = config.NewConfig(programName)

	// Metrics have to be registered to be exposed:
	//prometheus.MustRegister(minerTotalHashRate)
	prometheus.MustRegister(minerGpuHashRate)
}

type pmInfoFile struct {
	gpu  int
	path string
}

func mapRegexp(text string, expression string) []map[string]string {
	var res []map[string]string

	r := regexp.MustCompile(expression)

	subexpNames := r.SubexpNames()

	for _, submatchList := range r.FindAllStringSubmatch(text, -1) {
		var m = make(map[string]string)
		for i, submatch := range submatchList[1:] {
			m[subexpNames[i+1]] = submatch
		}
		res = append(res, m)
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
	)

	//
	// Find all amd_gpu_pm_info files
	//
	filepath.Walk("/sys/kernel/debug/dri", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "amdgpu_pm_info" {
			dir, _ := filepath.Split(path)
			dirList := strings.Split(strings.Trim(dir, "/"), "/")
			gpuString := dirList[len(dirList)-1]
			gpu, _ := strconv.Atoi(gpuString)

			pmInfoFileList = append(pmInfoFileList, pmInfoFile{gpu: gpu, path: path})
		}
		return nil
	})

	for _, info := range pmInfoFileList {
		log.Printf("gpu %d, path %s", info.gpu, info.path)

		bytes, err := ioutil.ReadFile(info.path)
		if err == nil {
			text := string(bytes)

			fmt.Println(text)

			res := mapRegexp(text, expressions[clock])
			fmt.Printf("list %v\n", res)

			res = mapRegexp(text, expressions[power])
			fmt.Printf("list %v\n", res)

			res = mapRegexp(text, expressions[temp])
			fmt.Printf("temp %v\n", res[0]["val"])

			res = mapRegexp(text, expressions[load])
			fmt.Printf("load %v\n", res[0]["val"])

			//r2 := regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?) W \(max GPU\)`)
			//fmt.Println(r2.FindStringSubmatch(text))

			//r3 := regexp.MustCompile(`(?P<value>[0-9]+(?:\.[0-9]+)?) W \((?P<stat>(?:[A-Za-z0-9\ ]+))\)`)
			//fmt.Println(r3.FindAllStringSubmatch(text, -1))
			//fmt.Println(r3.SubexpNames())

			//r4 := regexp.MustCompile(strings.Join(expressions, "|"))
			//fmt.Println(r4.FindAllStringSubmatch(text, -1))
			//fmt.Println(r4.SubexpNames())

		} else {
			log.Printf("Error reading file %v: %v\n", info.path, err)
		}
	}

	//go func() {
	//for {

	//minerGpuHashRate.With(prometheus.Labels{
	//"miner":  cfg.Miner.Program(),
	//"gpu":    fmt.Sprintf("GPU%d", device.ID),
	//"symbol": cfg.Miner.Symbol(),
	//}).Set(device.MHS20S)
	//} else {
	//log.Printf("Error connecting to miner: %v\n", err)
	//}

	//time.Sleep(time.Second * cfg.QueryDelay())
	//}
	//}()

	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	//http.Handle("/metrics", promhttp.Handler())
	//log.Fatal(http.ListenAndServe(cfg.Prometheus.Address(), nil))
}
