package main

import (
	"flag"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var addr = flag.String("listen-address", ":9776", "The address to listen on for HTTP requests.")
var enableRuntimeMetrics = flag.Bool("runtime-metrics", false, "Enable prometheus runtime metrics.")
var password = flag.String("password", "", "The password used for logging into S-Miles Cloud.")
var username = flag.String("username", "", "The username used for logging into S-Miles Cloud.")

var log = initLog()

func initLog() *zap.Logger {
	stdout := zapcore.AddSync(os.Stdout)
	config := zap.NewProductionEncoderConfig()
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	consoleEncoder := zapcore.NewConsoleEncoder(config)
	logCore := zapcore.NewCore(consoleEncoder, stdout, zap.DebugLevel)
	logger := zap.New(logCore)

	return logger
}

func main() {
	defer log.Sync()

	mainLog := log.Sugar().Named("main")

	flag.Parse()

	if len(*username) == 0 {
		mainLog.Fatal("Username must not be empty.")
	}
	if len(*password) == 0 {
		mainLog.Fatal("Password must not be empty.")
	}

	mainLog.Debug("Registering Prometheus metrics …")

	reg := prometheus.NewRegistry()

	reg.MustRegister(collectors.NewBuildInfoCollector())

	if *enableRuntimeMetrics {
		reg.MustRegister(collectors.NewGoCollector(
			collectors.WithGoCollectorRuntimeMetrics(collectors.GoRuntimeMetricsRule{Matcher: regexp.MustCompile("/.*")}),
		))
	}

	reg.MustRegister(newMetrics())

	mainLog.Infof("Listening for HTTP requests on %s …", *addr)

	prometheusLog, _ := zap.NewStdLogAt(log.Named("prometheus"), zap.ErrorLevel)

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			ErrorLog: prometheusLog,
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))
	mainLog.Fatal(http.ListenAndServe(*addr, nil))
}
