package middleware

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

type (
	PrometheusConfig struct {
		Skipper   middleware.Skipper
		Namespace string
	}
)

var (
	DefaultPrometheusConfig = PrometheusConfig{
		Skipper:   middleware.DefaultSkipper,
		Namespace: "whalefs",
	}
)

var (
	echoReqQps      *prometheus.CounterVec
	echoReqDuration *prometheus.SummaryVec
)

func initCollector(namespace string) {
	echoReqQps = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_request_total",
			Help:      "HTTP requests processed.",
		},
		[]string{"status"},
	)
	echoReqDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request latencies in seconds.",
		},
		[]string{},
	)
	prometheus.MustRegister(echoReqQps, echoReqDuration)
}

func NewMetric() echo.MiddlewareFunc {
	return NewMetricWithConfig(DefaultPrometheusConfig)
}

func NewMetricWithConfig(config PrometheusConfig) echo.MiddlewareFunc {
	initCollector(config.Namespace)
	if config.Skipper == nil {
		config.Skipper = DefaultPrometheusConfig.Skipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			res := c.Response()
			start := time.Now()

			if err := next(c); err != nil {
				c.Error(err)
			}
			status := strconv.Itoa(res.Status)
			elapsed := time.Since(start).Seconds()
			echoReqQps.WithLabelValues(status).Inc()
			echoReqDuration.WithLabelValues().Observe(elapsed)
			return nil
		}
	}
}
