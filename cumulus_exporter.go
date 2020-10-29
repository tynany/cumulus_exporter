package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"github.com/tynany/cumulus_exporter/collector"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9365").String()
	telemetryPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
)

func handler(w http.ResponseWriter, r *http.Request) {
	registry := prometheus.NewRegistry()

	registry.Register(collector.NewExporter())

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
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("cumulus_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
}

func main() {
	prometheus.MustRegister(version.NewCollector("cumulus_exporter"))

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
