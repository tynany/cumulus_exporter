package collector

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	resourceSubsystem = "resource"

	resourceLabels = []string{"resource"}

	resourceDesc = map[string]*prometheus.Desc{
		"max":  colPromDesc(resourceSubsystem, "maximum", "Resource Maximum.", resourceLabels),
		"used": colPromDesc(resourceSubsystem, "used", "Resource Used.", resourceLabels),
	}

	totalResourceErrors = 0.0

	cumulusCLResPath = kingpin.Flag("cumulus.cl-resource-query.path", "Path of cl-resource-query.").Default("/usr/cumulus/bin/cl-resource-query").String()
)

func init() {
	registerCollector(resourceSubsystem, enabledByDefault, NewResourceCollector)
}

// ResourceCollector collects resource metrics, implemented as per the Collector interface.
type ResourceCollector struct{}

// NewResourceCollector returns a new ResourceCollector.
func NewResourceCollector() Collector {
	return &ResourceCollector{}
}

// Get metrics and send to the Prometheus.Metric channel.
func (c *ResourceCollector) Get(ch chan<- prometheus.Metric) (float64, error) {

	jsonResources, err := getResourceStats()
	if err != nil {
		totalResourceErrors++
		return totalResourceErrors, fmt.Errorf("cannot get resources: %s", err)
	}
	if err := processResourceStats(ch, jsonResources); err != nil {
		totalResourceErrors++
		return totalResourceErrors, err
	}
	return totalResourceErrors, nil

}

func getResourceStats() ([]byte, error) {
	args := []string{"-j"}
	output, err := exec.Command(*cumulusCLResPath, args...).Output()
	if err != nil {
		return nil, err
	}
	return output, nil
}

func processResourceStats(ch chan<- prometheus.Metric, jsonResourceSum []byte) error {
	dataMap := map[string]*resourceEntry{}
	if err:= json.Unmarshal(jsonResourceSum, &dataMap); err != nil {
		return fmt.Errorf("cannot unmarshal resource json: %s", err)
	}

	for _, v := range dataMap {
	        newGauge(ch, resourceDesc["max"], v.Max, strings.ToLower(v.Name))
	        newGauge(ch, resourceDesc["used"], v.Max, strings.ToLower(v.Name))
	}

	return nil
}


type resourceEntry struct {
	Count float64 `json:"count"`
	Max   float64 `json:"max"`
	Name  string  `json:"name"`
}
