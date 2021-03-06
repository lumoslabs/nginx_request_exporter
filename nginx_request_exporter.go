// Copyright 2016 Markus Lindenberg
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/mcuadros/go-syslog.v2"
)

const (
	app       = "nginx_request_exporter"
	namespace = "nginx_request"

	defaultListenAddr    = ":9147"
	defaultTelemetryPath = "/metrics"
	defaultSyslogAddr    = "127.0.0.1:9514"
)

var (
	e                       *echo.Echo
	defaultHistogramBuckets = []string{".005", ".01", ".025", ".05", ".1", ".25", ".5", "1", "2.5", "5", "10"}
)

var (
	exporter      = kingpin.New(app, "Gather metrics from nginx access logs over a syslog channel.")
	confPath      = exporter.Flag("config", "Path to config file.").Short('C').Envar("NGX_REQUEST_EXPORTER_CONFIG_PATH").Required().ExistingFile()
	listen        = exporter.Flag("listen-address", "Address to listen on for scrapes.").Short('l').Default(defaultListenAddr).Envar("NGX_REQUEST_EXPORTER_LISTEN_ADDRESS").String()
	telmPath      = exporter.Flag("telemetry-path", "Path for exposing metrics.").Short('p').Default(defaultTelemetryPath).Envar("NGX_REQUEST_EXPORTER_TELEMETRY_PATH").String()
	syslogAddress = exporter.Flag("syslog-address", "Address for syslog.").Default(defaultSyslogAddr).Envar("NGZ_REQUEST_EXPORTER_SYSLOG_ADDRESS").String()
	metricBuckets = exporter.Flag("buckets", "Buckets for histogram.").Default(defaultHistogramBuckets...).Envar("NGX_REQUEST_EXPORTER_BUCKETS").Float64List()
	gzip          = exporter.Flag("gzip", "Use gzip compression on metrics response.").Envar("NGX_REQUEST_EXPORTER_GZIP").Bool()
	grace         = exporter.Flag("graceful-timeout", "Timeout for graceful shutdown.").Default("10s").Envar("NGX_REQUEST_EXPORTER_GRACEFUL_TIMEOUT").Duration()
	v             = exporter.Flag("v", "Log level. 0 = off, 1 = error, 2 = warn, 3 = info, 4 = debug").Short('v').Default("0").Envar("NGX_REQUEST_EXPORTER_LOG_LEVEL").Int()
)

func logLevel() (l log.Lvl) {
	switch *v {
	default:
		l = log.DEBUG
	case 0:
		l = log.OFF
	case 1:
		l = log.ERROR
	case 2:
		l = log.WARN
	case 3:
		l = log.INFO
	}
	return
}

func main() {
	exporter.Version(version())
	kingpin.MustParse(exporter.Parse(os.Args[1:]))
	cfg, er := Configure(*confPath, &Config{
		ListenAddress: *listen,
		TelemetryPath: *telmPath,
		SyslogAddress: *syslogAddress,
		Buckets:       *metricBuckets,
		Prefix: &LabelConfig{
			Default: "",
			Rules:   nil,
		},
		DeviceType: &LabelConfig{
			Default: "",
			Rules:   nil,
		},
	})
	if er != nil {
		panic(er)
	}

	e = echo.New()
	e.HideBanner = true
	e.Logger.SetLevel(logLevel())

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	if *gzip {
		e.Use(middleware.Gzip())
	}

	// Setup HTTP server
	e.GET(cfg.TelemetryPath, echo.WrapHandler(promhttp.Handler()))

	// Listen to signals
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Set up syslog server
	channel := make(syslog.LogPartsChannel, 20000)
	handler := syslog.NewChannelHandler(channel)
	server := syslog.NewServer()
	server.SetFormat(syslog.RFC3164)
	server.SetHandler(handler)

	var err error
	if strings.HasPrefix(cfg.SyslogAddress, "unix:") {
		err = server.ListenUnixgram(strings.TrimPrefix(cfg.SyslogAddress, "unix:"))
	} else {
		err = server.ListenUDP(cfg.SyslogAddress)
	}
	if err != nil {
		log.Fatal(err)
	}
	err = server.Boot()
	if err != nil {
		e.Logger.Fatal(err)
	}

	// Setup metrics
	syslogMessages := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "exporter_syslog_messages",
		Help:      "Current total syslog messages received.",
	})
	err = prometheus.Register(syslogMessages)
	if err != nil {
		log.Fatal(err)
	}
	syslogParseFailures := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "exporter_syslog_parse_failure",
		Help:      "Number of errors while parsing syslog messages.",
	})
	err = prometheus.Register(syslogParseFailures)
	if err != nil {
		e.Logger.Fatal(err)
	}
	var msgs int64
	go func() {
		for part := range channel {
			syslogMessages.Inc()
			msgs++
			tag, _ := part["tag"].(string)
			if tag != "nginx" {
				e.Logger.Warn("Ignoring syslog message with wrong tag")
				syslogParseFailures.Inc()
				continue
			}
			server, _ := part["hostname"].(string)
			if server == "" {
				e.Logger.Warn("Hostname missing in syslog message")
				syslogParseFailures.Inc()
				continue
			}

			content, _ := part["content"].(string)
			if content == "" {
				e.Logger.Warn("Ignoring empty syslog message")
				syslogParseFailures.Inc()
				continue
			}

			metrics, labels, err := parseMessage(content)
			if err != nil {
				e.Logger.Error(err)
				continue
			}

			// Lumos magic: get device_type from http user agent
			if userAgent, ok := labels.Get("user_agent"); ok && cfg.DeviceType != nil {
				labels.Set("device_type", parseRule(userAgent, cfg.DeviceType.Default, cfg.DeviceType.Rules))
			}
			labels.Delete("user_agent")

			// Lumos magic: get prefix from request uri
			if requestURI, ok := labels.Get("request_uri"); ok && cfg.Prefix != nil {
				labels.Set("prefix", parseRule(requestURI, cfg.Prefix.Default, cfg.Prefix.Rules))
			}
			labels.Delete("request_uri")

			var collector prometheus.Collector

			for _, metric := range metrics {
				// Create histogram metrics if there are any matches according to histogram rules
				if histos, ok := parseHistograms(metric.Name, labels, cfg.HistogramRules); ok {
					for _, histo := range histos {
						collector = prometheus.NewHistogramVec(prometheus.HistogramOpts{
							Namespace: namespace,
							Name:      histo.Name,
							Help:      fmt.Sprintf("Nginx request log value for %s", metric.Name),
							Buckets:   cfg.Buckets,
						}, histo.Labels.Names)
						if err := prometheus.Register(collector); err != nil {
							if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
								collector = are.ExistingCollector.(*prometheus.HistogramVec)
							} else {
								log.Error(err)
								continue
							}
						}
						collector.(*prometheus.HistogramVec).WithLabelValues(histo.Labels.Values...).Observe(metric.Value)
					}
				}
			}

			// Always create a counter
			collector = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: namespace,
				Name:      fmt.Sprint("requests_total"),
				Help:      fmt.Sprintf("Nginx total requests"),
			}, labels.Names)
			if err := prometheus.Register(collector); err != nil {
				if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
					collector = are.ExistingCollector.(*prometheus.CounterVec)
				} else {
					log.Error(err)
					continue
				}
			}
			collector.(*prometheus.CounterVec).WithLabelValues(labels.Values...).Inc()
		}
	}()

	go func() {
		e.Logger.Infof("Starting nginx_request_exporter version=%v address=%v", version(), cfg.ListenAddress)
		if er := e.Start(cfg.ListenAddress); err != nil {
			e.Logger.Error(er)
		}
	}()

	<-sigchan
	e.Logger.Info("Shutting down the server...")
	if er := server.Kill(); er != nil {
		e.Logger.Error(er)
	}
	ctx, cancel := context.WithTimeout(context.Background(), *grace)
	defer cancel()
	if er := e.Shutdown(ctx); er != nil {
		e.Logger.Fatal(er)
	}
}
