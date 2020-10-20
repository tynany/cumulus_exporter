package collector

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	sensorSubsystem = "sensor"

	sensorLabels = []string{"sensor", "description"}

	sensorDesc = map[string]*prometheus.Desc{
		"state":       colPromDesc(sensorSubsystem, "state", "State of Sensor (0 = Bad, 1 = Ok, 2 = Absent).", sensorLabels),
		"temp":        colPromDesc(sensorSubsystem, "temperature_celsius", "Temperature in Celsius.", sensorLabels),
		"fanSpeed":    colPromDesc(sensorSubsystem, "fan_speed_rpm", "Fan Speed RPM.", sensorLabels),
		"minTemp":     colPromDesc(sensorSubsystem, "minimum_operating_temperature_celsius", "Minimum Operating Temperature in Celsius.", sensorLabels),
		"maxTemp":     colPromDesc(sensorSubsystem, "maximum_operating_temperature_celsius", "Maximum Operating Temperature in Celsius.", sensorLabels),
		"minFanSpeed": colPromDesc(sensorSubsystem, "minimum_operating_fan_speed_rpm", "Minimum Operating Fan Speed RPM.", sensorLabels),
		"maxFanSpeed": colPromDesc(sensorSubsystem, "maximum_operating_fan_speed_rpm", "Maximum Operating Fan Speed RPM.", sensorLabels),
	}

	sensorErrors      = []error{}
	totalSensorErrors = 0.0
)

// SensorCollector collects Sensor metrics, implemented as per prometheus.Collector interface.
type SensorCollector struct{}

// NewSensorCollector returns a SensorCollector struct.
func NewSensorCollector() *SensorCollector {
	return &SensorCollector{}
}

// Name of the collector. Used to populate flag name.
func (*SensorCollector) Name() string {
	return sensorSubsystem
}

// Help describes the metrics this collector scrapes. Used to populate flag help.
func (*SensorCollector) Help() string {
	return "Collect Sensor Metrics"
}

// EnabledByDefault describes whether this collector is enabled by default. Used to populate flag default.
func (*SensorCollector) EnabledByDefault() bool {
	return true
}

// Describe implemented as per the prometheus.Collector interface.
func (*SensorCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range sensorDesc {
		ch <- desc
	}
}

// Collect implemented as per the prometheus.Collector interface.
func (c *SensorCollector) Collect(ch chan<- prometheus.Metric) {

	jsonSensors, err := getSensorStats()
	if err != nil {
		totalSensorErrors++
		sensorErrors = append(sensorErrors, fmt.Errorf("cannot get sensors: %s", err))
	} else {
		if err := processSensorStats(ch, jsonSensors); err != nil {
			totalSensorErrors++
			sensorErrors = append(sensorErrors, err)
		}
	}
}

// CollectErrors returns what errors have been gathered.
func (*SensorCollector) CollectErrors() []error {
	errors := sensorErrors
	sensorErrors = []error{}
	return errors
}

// CollectTotalErrors returns total errors.
func (*SensorCollector) CollectTotalErrors() float64 {
	return totalSensorErrors
}

func getSensorStats() ([]byte, error) {
	args := []string{"-j"}
	output, err := exec.Command(smonctlPath, args...).Output()
	if err != nil {
		return nil, err
	}
	return output, nil
}

func processSensorStats(ch chan<- prometheus.Metric, jsonSensorSum []byte) error {
	var jsonSensors sensorData

	if err := json.Unmarshal(jsonSensorSum, &jsonSensors); err != nil {
		return fmt.Errorf("cannot unmarshal sensor json: %s", err)
	}

	for _, sensor := range jsonSensors {
		labels := []string{strings.ToLower(sensor.Name), strings.ToLower(sensor.Description)}

		if strings.ToLower(sensor.State) == "ok" {
			newGauge(ch, sensorDesc["state"], 1.0, labels...)

			if sensor.Type == "fan" {
				newGauge(ch, sensorDesc["fanSpeed"], sensor.Input, labels...)
				newGauge(ch, sensorDesc["minFanSpeed"], sensor.Min, labels...)
				newGauge(ch, sensorDesc["maxFanSpeed"], sensor.Max, labels...)
			}
			if sensor.Type == "temp" {
				newGauge(ch, sensorDesc["temp"], sensor.Input, labels...)
				newGauge(ch, sensorDesc["minTemp"], sensor.Min, labels...)
				newGauge(ch, sensorDesc["maxTemp"], sensor.Max, labels...)
			}

		} else if strings.ToLower(sensor.State) == "absent" {
			newGauge(ch, sensorDesc["state"], 2.0, labels...)
		} else {
			newGauge(ch, sensorDesc["state"], 0.0, labels...)
		}
	}
	return nil
}

type sensorData []struct {
	Name  string  `json:"name"`
	Description string `json:"description"`
	State string  `json:"state"`
	Input float64 `json:"input,omitempty"`
	Type  string  `json:"type"`
	Max   float64 `json:"max,omitempty"`
	Min   float64 `json:"min,omitempty"`
}
