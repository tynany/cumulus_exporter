package collector

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	resourceSubsystem = "resource"

	resourceLabels = []string{"resource"}

	resourceDesc = map[string]*prometheus.Desc{
		"max":  colPromDesc(resourceSubsystem, "maximum", "Resource Maximum.", resourceLabels),
		"used": colPromDesc(resourceSubsystem, "used", "Resource Used.", resourceLabels),
	}

	resourceErrors      = []error{}
	totalResourceErrors = 0.0
)

// ResourceCollector collects Resource metrics, implemented as per prometheus.Collector interface.
type ResourceCollector struct{}

// NewResourceCollector returns a ResourceCollector struct.
func NewResourceCollector() *ResourceCollector {
	return &ResourceCollector{}
}

// Name of the collector. Used to populate flag name.
func (*ResourceCollector) Name() string {
	return resourceSubsystem
}

// Help describes the metrics this collector scrapes. Used to populate flag help.
func (*ResourceCollector) Help() string {
	return "Collect Resource Metrics"
}

// EnabledByDefault describes whether this collector is enabled by default. Used to populate flag default.
func (*ResourceCollector) EnabledByDefault() bool {
	return true
}

// Describe implemented as per the prometheus.Collector interface.
func (*ResourceCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range resourceDesc {
		ch <- desc
	}
}

// Collect implemented as per the prometheus.Collector interface.
func (c *ResourceCollector) Collect(ch chan<- prometheus.Metric) {

	jsonResources, err := getResourceStats()
	if err != nil {
		totalResourceErrors++
		resourceErrors = append(resourceErrors, fmt.Errorf("cannot get resources: %s", err))
	} else {
		if err := processResourceStats(ch, jsonResources); err != nil {
			totalResourceErrors++
			resourceErrors = append(resourceErrors, err)
		}
	}
}

// CollectErrors returns what errors have been gathered.
func (*ResourceCollector) CollectErrors() []error {
	errors := resourceErrors
	resourceErrors = []error{}
	return errors
}

// CollectTotalErrors returns total errors.
func (*ResourceCollector) CollectTotalErrors() float64 {
	return totalResourceErrors
}

func getResourceStats() ([]byte, error) {
	args := []string{"-j"}
	output, err := exec.Command(clResourceQueryPath, args...).Output()
	if err != nil {
		return nil, err
	}
	return output, nil
}

func processResourceStats(ch chan<- prometheus.Metric, jsonResourceSum []byte) error {
	var jsonResources resourceData

	if err := json.Unmarshal(jsonResourceSum, &jsonResources); err != nil {
		return fmt.Errorf("cannot unmarshal resource json: %s", err)
	}

	newGauge(ch, resourceDesc["max"], jsonResources.ACLL4PortRangeCheckers.Max, strings.ToLower(jsonResources.ACLL4PortRangeCheckers.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.ACLL4PortRangeCheckers.Count, strings.ToLower(jsonResources.ACLL4PortRangeCheckers.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.ECMPNhEntry.Max, strings.ToLower(jsonResources.ECMPNhEntry.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.ECMPNhEntry.Count, strings.ToLower(jsonResources.ECMPNhEntry.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.EgACLCounter.Max, strings.ToLower(jsonResources.EgACLCounter.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.EgACLCounter.Count, strings.ToLower(jsonResources.EgACLCounter.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.EgACLEntry.Max, strings.ToLower(jsonResources.EgACLEntry.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.EgACLEntry.Count, strings.ToLower(jsonResources.EgACLEntry.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.EgACLMeter.Max, strings.ToLower(jsonResources.EgACLMeter.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.EgACLMeter.Count, strings.ToLower(jsonResources.EgACLMeter.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.EgACLSlice.Max, strings.ToLower(jsonResources.EgACLSlice.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.EgACLSlice.Count, strings.ToLower(jsonResources.EgACLSlice.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.EgACLV4MacFilter.Max, strings.ToLower(jsonResources.EgACLV4MacFilter.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.EgACLV4MacFilter.Count, strings.ToLower(jsonResources.EgACLV4MacFilter.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.EgACLV6Filter.Max, strings.ToLower(jsonResources.EgACLV6Filter.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.EgACLV6Filter.Count, strings.ToLower(jsonResources.EgACLV6Filter.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.Host0Entry.Max, strings.ToLower(jsonResources.Host0Entry.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.Host0Entry.Count, strings.ToLower(jsonResources.Host0Entry.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.HostV4Entry.Max, strings.ToLower(jsonResources.HostV4Entry.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.HostV4Entry.Count, strings.ToLower(jsonResources.HostV4Entry.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.HostV6Entry.Max, strings.ToLower(jsonResources.HostV6Entry.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.HostV6Entry.Count, strings.ToLower(jsonResources.HostV6Entry.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.InACL8021XFilter.Max, strings.ToLower(jsonResources.InACL8021XFilter.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.InACL8021XFilter.Count, strings.ToLower(jsonResources.InACL8021XFilter.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.InACLCounter.Max, strings.ToLower(jsonResources.InACLCounter.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.InACLCounter.Count, strings.ToLower(jsonResources.InACLCounter.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.InACLEntry.Max, strings.ToLower(jsonResources.InACLEntry.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.InACLEntry.Count, strings.ToLower(jsonResources.InACLEntry.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.InACLMeter.Max, strings.ToLower(jsonResources.InACLMeter.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.InACLMeter.Count, strings.ToLower(jsonResources.InACLMeter.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.InACLMirrorFilter.Max, strings.ToLower(jsonResources.InACLMirrorFilter.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.InACLMirrorFilter.Count, strings.ToLower(jsonResources.InACLMirrorFilter.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.InACLSlice.Max, strings.ToLower(jsonResources.InACLSlice.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.InACLSlice.Count, strings.ToLower(jsonResources.InACLSlice.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.InACLV4MacFilter.Max, strings.ToLower(jsonResources.InACLV4MacFilter.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.InACLV4MacFilter.Count, strings.ToLower(jsonResources.InACLV4MacFilter.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.InACLV4MacMangle.Max, strings.ToLower(jsonResources.InACLV4MacMangle.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.InACLV4MacMangle.Count, strings.ToLower(jsonResources.InACLV4MacMangle.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.InACLV6Filter.Max, strings.ToLower(jsonResources.InACLV6Filter.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.InACLV6Filter.Count, strings.ToLower(jsonResources.InACLV6Filter.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.InACLV6Mangle.Max, strings.ToLower(jsonResources.InACLV6Mangle.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.InACLV6Mangle.Count, strings.ToLower(jsonResources.InACLV6Mangle.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.InPbrV4MacFilter.Max, strings.ToLower(jsonResources.InPbrV4MacFilter.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.InPbrV4MacFilter.Count, strings.ToLower(jsonResources.InPbrV4MacFilter.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.InPbrV6Filter.Max, strings.ToLower(jsonResources.InPbrV6Filter.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.InPbrV6Filter.Count, strings.ToLower(jsonResources.InPbrV6Filter.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.MacEntry.Max, strings.ToLower(jsonResources.MacEntry.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.MacEntry.Count, strings.ToLower(jsonResources.MacEntry.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.MrouteTotalEntry.Max, strings.ToLower(jsonResources.MrouteTotalEntry.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.MrouteTotalEntry.Count, strings.ToLower(jsonResources.MrouteTotalEntry.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.Route0Entry.Max, strings.ToLower(jsonResources.Route0Entry.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.Route0Entry.Count, strings.ToLower(jsonResources.Route0Entry.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.Route1Entry.Max, strings.ToLower(jsonResources.Route1Entry.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.Route1Entry.Count, strings.ToLower(jsonResources.Route1Entry.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.RouteTotalEntry.Max, strings.ToLower(jsonResources.RouteTotalEntry.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.RouteTotalEntry.Count, strings.ToLower(jsonResources.RouteTotalEntry.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.RouteV4Entry.Max, strings.ToLower(jsonResources.RouteV4Entry.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.RouteV4Entry.Count, strings.ToLower(jsonResources.RouteV4Entry.Name))

	newGauge(ch, resourceDesc["max"], jsonResources.RouteV6Entry.Max, strings.ToLower(jsonResources.RouteV6Entry.Name))
	newGauge(ch, resourceDesc["used"], jsonResources.RouteV6Entry.Count, strings.ToLower(jsonResources.RouteV6Entry.Name))

	return nil
}

type resourceData struct {
	ACLL4PortRangeCheckers resourceEntry `json:"acl_l4_port_range_checkers"`
	ECMPNhEntry            resourceEntry `json:"ecmp_nh_entry"`
	EgACLCounter           resourceEntry `json:"eg_acl_counter"`
	EgACLEntry             resourceEntry `json:"eg_acl_entry"`
	EgACLMeter             resourceEntry `json:"eg_acl_meter"`
	EgACLSlice             resourceEntry `json:"eg_acl_slice"`
	EgACLV4MacFilter       resourceEntry `json:"eg_acl_v4mac_filter"`
	EgACLV6Filter          resourceEntry `json:"eg_acl_v6_filter"`
	Host0Entry             resourceEntry `json:"host_0_entry"`
	HostV4Entry            resourceEntry `json:"host_v4_entry"`
	HostV6Entry            resourceEntry `json:"host_v6_entry"`
	InACL8021XFilter       resourceEntry `json:"in_acl_8021x_filter"`
	InACLCounter           resourceEntry `json:"in_acl_counter"`
	InACLEntry             resourceEntry `json:"in_acl_entry"`
	InACLMeter             resourceEntry `json:"in_acl_meter"`
	InACLMirrorFilter      resourceEntry `json:"in_acl_mirror_filter"`
	InACLSlice             resourceEntry `json:"in_acl_slice"`
	InACLV4MacFilter       resourceEntry `json:"in_acl_v4mac_filter"`
	InACLV4MacMangle       resourceEntry `json:"in_acl_v4mac_mangle"`
	InACLV6Filter          resourceEntry `json:"in_acl_v6_filter"`
	InACLV6Mangle          resourceEntry `json:"in_acl_v6_mangle"`
	InPbrV4MacFilter       resourceEntry `json:"in_pbr_v4mac_filter"`
	InPbrV6Filter          resourceEntry `json:"in_pbr_v6_filter"`
	MacEntry               resourceEntry `json:"mac_entry"`
	MrouteTotalEntry       resourceEntry `json:"mroute_total_entry"`
	Route0Entry            resourceEntry `json:"route_0_entry"`
	Route1Entry            resourceEntry `json:"route_1_entry"`
	RouteTotalEntry        resourceEntry `json:"route_total_entry"`
	RouteV4Entry           resourceEntry `json:"route_v4_entry"`
	RouteV6Entry           resourceEntry `json:"route_v6_entry"`
}

type resourceEntry struct {
	Count float64 `json:"count"`
	Max   float64 `json:"max"`
	Name  string  `json:"name"`
}
