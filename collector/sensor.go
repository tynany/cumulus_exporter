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

	totalSensorErrors = 0.0

	cumulusSMONCTLPath = kingpin.Flag("cumulus.smonctl.path", "Path of smonctl.").Default("/usr/sbin/smonctl").String()
)

func init() {
	registerCollector(sensorSubsystem, enabledByDefault, NewSensorCollector)
}

// SensorCollector collects sensor metrics, implemented as per the Collector interface.
type SensorCollector struct{}

// NewSensorCollector returns a new SensorCollector.
func NewSensorCollector() Collector {
	return &SensorCollector{}
}

// Get metrics and send to the Prometheus.Metric channel.
func (c *SensorCollector) Get(ch chan<- prometheus.Metric) (float64, error) {
	jsonSensors, err := getSensorStats()
	if err != nil {
		totalSensorErrors++
		return totalSensorErrors, fmt.Errorf("cannot get sensors: %s", err)
	}
	if err := processSensorStats(ch, jsonSensors); err != nil {
		totalSensorErrors++
		return totalSensorErrors, err
	}
	return totalSensorErrors, nil
}

func getSensorStats() ([]byte, error) {
	args := []string{"-j"}
	output, err := exec.Command(*cumulusSMONCTLPath, args...).Output()
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
	Name        string  `json:"name"`
	Description string  `json:"description"`
	State       string  `json:"state"`
	Input       float64 `json:"input,omitempty"`
	Type        string  `json:"type"`
	Max         float64 `json:"max,omitempty"`
	Min         float64 `json:"min,omitempty"`
}
