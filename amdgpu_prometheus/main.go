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
	"time"

	"../config"
	"../export"
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
		cfg            *config.Config  = config.NewConfig(programName)
		exporter       export.Exporter = export.NewPrometheus(cfg.Prometheus.Address())
	)

	exporter = export.NewPrometheus(cfg.Prometheus.Address())

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

					for _, ctype := range []string{clock, power, temp, load} {
						for name, value := range mapRegexp(text, expressions[ctype]) {
							exporter.Export(ctype, gpu, name, value)
						}
					}
				} else {
					log.Printf("Error reading file %v: %v\n", info.path, err)
				}
			}

			time.Sleep(time.Second * cfg.QueryDelay())
		}
	}()

	exporter.Setup()
}
