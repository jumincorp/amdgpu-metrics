package export

const (
	clock = "clock"
	power = "power"
	temp  = "temp"
	load  = "load"
)

type Exporter interface {
	Setup()
	Export(ctype string, gpu string, name string, value string) error
}
