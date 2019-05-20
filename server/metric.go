package server

import (
	"math/rand"
	"time"

	"github.com/labstack/echo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Namespace:   "whalefs",
		Name:        "myapp_processed_ops_total",
		Help:        "The total number of processed events",
		ConstLabels: map[string]string{"name": "system", "mode": "upload"},
	})
	qpsProcessed = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace:   "whalefs",
		Subsystem:   "download",
		Name:        "allocate_bytes",
		ConstLabels: map[string]string{"name": "system", "mode": "upload"},
	})
)

func init() {
	go func() {
		for {
			opsProcessed.Inc()
			qpsProcessed.Set(rand.Float64())
			time.Sleep(2 * time.Second)
		}
	}()
}

func (s *Server) metric(ctx echo.Context) error {
	handler := promhttp.Handler()
	handler.ServeHTTP(ctx.Response(), ctx.Request())
	return nil
}
