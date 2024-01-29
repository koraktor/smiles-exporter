package main

import (
	"flag"
	"log"
	"net/http"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
var enableRuntimeMetrics = flag.Bool("runtime-metrics", false, "Enable prometheus runtime metrics.")
var password = flag.String("password", "", "The password used for logging into S-Miles Cloud.")
var username = flag.String("username", "", "The username used for logging into S-Miles Cloud.")

func main() {
	flag.Parse()

	reg := prometheus.NewRegistry()

	reg.MustRegister(collectors.NewBuildInfoCollector())

	if *enableRuntimeMetrics {
		reg.MustRegister(collectors.NewGoCollector(
			collectors.WithGoCollectorRuntimeMetrics(collectors.GoRuntimeMetricsRule{Matcher: regexp.MustCompile("/.*")}),
		))
	}

	reg.MustRegister(newMetrics())

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))
	log.Fatal(http.ListenAndServe(*addr, nil))
}
