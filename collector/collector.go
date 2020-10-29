package collector

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	// The namespace used by all metrics.
	namespace = "cumulus"

	enabledByDefault  = true
	disabledByDefault = false
)

var (
	cumulusTotalScrapeCount = 0.0

	cumulusLabels = []string{"collector"}
	cumulusDesc   = map[string]*prometheus.Desc{
		"scrapesTotal":   promDesc("scrapes_total", "Total number of times cumulus_exporter has been scraped.", nil),
		"scrapeErrTotal": promDesc("scrape_errors_total", "Total number of errors from a collector.", cumulusLabels),
		"scrapeDuration": promDesc("scrape_duration_seconds", "Time it took for a collector's scrape to complete.", cumulusLabels),
		"collectorUp":    promDesc("collector_up", "Whether the collector's last scrape was successful (1 = successful, 0 = unsuccessful).", cumulusLabels),
	}

	allCollectors  = make(map[string]func() Collector)
	collectorState = make(map[string]*bool)
)

func registerCollector(name string, enabledByDefault bool, collector func() Collector) {
	defaultState := "disabled"
	if enabledByDefault {
		defaultState = "enabled"
	}

	allCollectors[name] = collector
	collectorState[name] = kingpin.Flag(fmt.Sprintf("collector.%s", name), fmt.Sprintf("Enable the %s collector (default: %s).", name, defaultState)).Default(strconv.FormatBool(enabledByDefault)).Bool()
}

// Collector is the interface a collector has to implement.
type Collector interface {
	// Gets metrics and sends to the Prometheus.Metric channel.
	Get(ch chan<- prometheus.Metric) (float64, error)
}

// Exporter collects all collector metrics, implemented as per the prometheus.Collector interface.
type Exporter struct {
	Collectors map[string]Collector
}

// NewExporter returns a new Exporter.
func NewExporter() *Exporter {
	enabledCollectors := make(map[string]Collector)
	for name, collector := range allCollectors {
		if *collectorState[name] {
			enabledCollectors[name] = collector()
		}
	}
	return &Exporter{
		Collectors: enabledCollectors,
	}
}

// Collect implemented as per the prometheus.Collector interface.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	cumulusTotalScrapeCount++
	ch <- prometheus.MustNewConstMetric(cumulusDesc["scrapesTotal"], prometheus.CounterValue, cumulusTotalScrapeCount)

	wg := &sync.WaitGroup{}
	for name, collector := range e.Collectors {
		wg.Add(1)
		go runCollector(ch, name, collector, wg)
	}
	wg.Wait()
}

func runCollector(ch chan<- prometheus.Metric, name string, collector Collector, wg *sync.WaitGroup) {
	defer wg.Done()

	startTime := time.Now()
	totalErrors, err := collector.Get(ch)

	ch <- prometheus.MustNewConstMetric(cumulusDesc["scrapeDuration"], prometheus.GaugeValue, float64(time.Since(startTime).Seconds()), name)
	ch <- prometheus.MustNewConstMetric(cumulusDesc["scrapeErrTotal"], prometheus.GaugeValue, totalErrors, name)

	if err != nil {
		ch <- prometheus.MustNewConstMetric(cumulusDesc["collectorUp"], prometheus.GaugeValue, 0, name)
		log.Errorf("collector %q scrape failed: %s", name, err)
	} else {
		ch <- prometheus.MustNewConstMetric(cumulusDesc["collectorUp"], prometheus.GaugeValue, 1, name)
	}

}

// Describe implemented as per the prometheus.Collector interface.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range cumulusDesc {
		ch <- desc
	}
}

func promDesc(metricName string, metricDescription string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(namespace+"_"+metricName, metricDescription, labels, nil)
}

func colPromDesc(subsystem string, metricName string, metricDescription string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, metricName), metricDescription, labels, nil)
}

func newGauge(ch chan<- prometheus.Metric, descName *prometheus.Desc, metric float64, labels ...string) {
	ch <- prometheus.MustNewConstMetric(descName, prometheus.GaugeValue, metric, labels...)
}

func newCounter(ch chan<- prometheus.Metric, descName *prometheus.Desc, metric float64, labels ...string) {
	ch <- prometheus.MustNewConstMetric(descName, prometheus.CounterValue, metric, labels...)
}
