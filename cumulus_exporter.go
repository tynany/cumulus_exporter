package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"github.com/tynany/cumulus_exporter/collector"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress      = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9365").String()
	telemetryPath      = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	cumulusSMONCTLPath = kingpin.Flag("cumulus.smonctl.path", "Path of smonctl.").Default("/usr/sbin/smonctl").String()
	cumulusCLResPath   = kingpin.Flag("cumulus.cl-resource-query.path", "Path of cl-resource-query.").Default("/usr/cumulus/bin/cl-resource-query").String()

	collectors = []*collector.Collector{}
)

func initCollectors() {
	sensor := collector.NewSensorCollector()
	collectors = append(collectors, &collector.Collector{
		Name:          sensor.Name(),
		PromCollector: sensor,
		Errors:        sensor,
		CLIHelper:     sensor,
	})
	resource := collector.NewResourceCollector()
	collectors = append(collectors, &collector.Collector{
		Name:          resource.Name(),
		PromCollector: resource,
		Errors:        resource,
		CLIHelper:     resource,
	})
}

func handler(w http.ResponseWriter, r *http.Request) {
	registry := prometheus.NewRegistry()
	enabledCollectors := []*collector.Collector{}
	for _, collector := range collectors {
		if *collector.Enabled {
			enabledCollectors = append(enabledCollectors, collector)
		}
	}
	ne := collector.NewExporter(enabledCollectors)
	ne.SetSMONCTLPath(*cumulusSMONCTLPath)
	ne.SetCLResPath(*cumulusCLResPath)
	registry.Register(ne)

	gatheres := prometheus.Gatherers{
		prometheus.DefaultGatherer,
		registry,
	}
	handlerOpts := promhttp.HandlerOpts{
		ErrorLog:      log.NewErrorLogger(),
		ErrorHandling: promhttp.ContinueOnError,
	}
	promhttp.HandlerFor(gatheres, handlerOpts).ServeHTTP(w, r)
}

func parseCLI() {
	for _, collector := range collectors {
		defaultState := "disabled"
		enabledByDefault := collector.CLIHelper.EnabledByDefault()
		if enabledByDefault == true {
			defaultState = "enabled"
		}
		collector.Enabled = kingpin.Flag(fmt.Sprintf("collector.%s", collector.CLIHelper.Name()), fmt.Sprintf("%s (default: %s).", collector.CLIHelper.Help(), defaultState)).Default(strconv.FormatBool(enabledByDefault)).Bool()
	}
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("cumulus_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
}

func main() {

	prometheus.MustRegister(version.NewCollector("cumulus_exporter"))

	initCollectors()
	parseCLI()

	log.Infof("Starting cumulus_exporter %s on %s", version.Info(), *listenAddress)

	http.HandleFunc(*telemetryPath, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Cumulus Exporter</title></head>
			<body>
			<h1>Cumulus Exporter</h1>
			<p><a href="` + *telemetryPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatal(err)
	}
}
