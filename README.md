# Cumulus Linux Exporter

Prometheus exporter for Cumulus Linux version 3.0+ that collects metrics and exposes them via HTTP, ready for collecting by Prometheus.

## Getting Started
To run cumulus_exporter:
```
./cumulus_exporter [flags]
```

To view metrics on the default port (9365) and path (/metrics):
```
http://device:9365/metrics
```

To view available flags:
```
usage: cumulus_exporter [<flags>]

Flags:
  -h, --help                Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":9365"
                            Address on which to expose metrics and web interface.
      --web.telemetry-path="/metrics"
                            Path under which to expose metrics.
      --cumulus.smonctl.path="/usr/sbin/smonctl"
                            Path of smonctl.
      --cumulus.cl-resource-query.path="/usr/cumulus/bin/cl-resource-query"
                            Path of cl-resource-query.
      --collector.sensor    Collect Sensor Metrics (default: enabled).
      --collector.resource  Collect Resource Metrics (default: enabled).
      --log.level="info"    Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]
      --log.format="logger:stderr"
                            Set the log target and format. Example: "logger:syslog?appname=bob&local=7" or "logger:stdout?json=true"
      --version             Show application version.
```

Promethues configuraiton:
```
scrape_configs:
  - job_name: cumulus
    static_configs:
      - targets:
        - device1:9365
        - device2:9365
    relabel_configs:
      - source_labels: [__address__]
        regex: "(.*):\d+"
        target: instance
```

## Collectors
To disable a default collector, use the `--no-collector.$name` flag.

### Enabled by Default
Name | Description
--- | ---
Sensors | Temperature, fan speed, etc.
Resources | Hardware resources utilisation, such as IPv4/IPv6 host entries, ACL entries, etc.
Version | Cumulus Linux version as identified in `/etc/lsb-release`.

## Development
### Building
```
go get github.com/tynany/cumulus_exporter
cd ${GOPATH}/src/github.com/prometheus/cumulus_exporter
go build
```
