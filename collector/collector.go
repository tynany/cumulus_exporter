package collector

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// The namespace used by all metrics.
const namespace = "cumulus"

var (
	cumulusTotalScrapeCount = 0.0

	cumulusLabels = []string{"collector"}
	cumulusDesc   = map[string]*prometheus.Desc{
		"scrapesTotal":   promDesc("scrapes_total", "Total number of times cumulus_exporter has been scraped.", nil),
		"scrapeErrTotal": promDesc("scrape_errors_total", "Total number of errors from a collector.", cumulusLabels),
		"scrapeDuration": promDesc("scrape_duration_seconds", "Time it took for a collector's scrape to complete.", cumulusLabels),
		"collectorUp":    promDesc("collector_up", "Whether the collector's last scrape was successful (1 = successful, 0 = unsuccessful).", cumulusLabels),
	}

	smonctlPath         string
	clResourceQueryPath string
)

// CLIHelper is used to populate flags.
type CLIHelper interface {
	// What the collector does.
	Help() string

	// Name of the collector.
	Name() string

	// Whether or not the collector is enabled by default.
	EnabledByDefault() bool
}

// CollectErrors is used to collect collector errors.
type CollectErrors interface {
	// Returns any errors that were encounted during Collect.
	CollectErrors() []error

	// Returns the total number of errors encounter during app run duration.
	CollectTotalErrors() float64
}

// Exporters contains a slice of Collectors.
type Exporters struct {
	Collectors []*Collector
}

// Collector contains everything needed to collect from a collector.
type Collector struct {
	Enabled       *bool
	Name          string
	PromCollector prometheus.Collector
	Errors        CollectErrors
	CLIHelper     CLIHelper
}

// NewExporter returns an Exporters type containing a slice of Collectors.
func NewExporter(collectors []*Collector) *Exporters {
	return &Exporters{Collectors: collectors}
}

// SetSMONCTLPath sets the path of smonctl.
func (e *Exporters) SetSMONCTLPath(path string) {
	smonctlPath = path
}

// SetCLResPath sets the path of cl-resource-query.
func (e *Exporters) SetCLResPath(path string) {
	clResourceQueryPath = path
}

// Describe implemented as per the prometheus.Collector interface.
func (e *Exporters) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range sensorDesc {
		ch <- desc
	}
	for _, collector := range e.Collectors {
		collector.PromCollector.Describe(ch)
	}
}

// Collect implemented as per the prometheus.Collector interface.
func (e *Exporters) Collect(ch chan<- prometheus.Metric) {
	cumulusTotalScrapeCount++
	ch <- prometheus.MustNewConstMetric(cumulusDesc["scrapesTotal"], prometheus.CounterValue, cumulusTotalScrapeCount)

	wg := &sync.WaitGroup{}
	for _, collector := range e.Collectors {
		wg.Add(1)
		go runCollector(ch, collector, wg)
	}
	wg.Wait()
}

func runCollector(ch chan<- prometheus.Metric, collector *Collector, wg *sync.WaitGroup) {
	defer wg.Done()
	startTime := time.Now()

	collector.PromCollector.Collect(ch)

	ch <- prometheus.MustNewConstMetric(cumulusDesc["scrapeErrTotal"], prometheus.GaugeValue, collector.Errors.CollectTotalErrors(), collector.Name)

	errors := collector.Errors.CollectErrors()
	if len(errors) > 0 {
		ch <- prometheus.MustNewConstMetric(cumulusDesc["collectorUp"], prometheus.GaugeValue, 0, collector.Name)
		for _, err := range errors {
			log.Errorf("collector \"%s\" scrape failed: %s", collector.Name, err)
		}
	} else {
		ch <- prometheus.MustNewConstMetric(cumulusDesc["collectorUp"], prometheus.GaugeValue, 1, collector.Name)
	}
	ch <- prometheus.MustNewConstMetric(cumulusDesc["scrapeDuration"], prometheus.GaugeValue, float64(time.Since(startTime).Seconds()), collector.Name)
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
