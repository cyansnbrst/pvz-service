package metric

import (
	"log"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// App metrics interface
type Metrics interface {
	IncHits(status int, method, path string)
	ObserveResponseTime(status int, method, path string, observeTime float64)
	IncPVZCreated()
	IncReceptionsCreated()
	IncProductsAdded()
}

// Prometheus metrics struct
type PrometheusMetrics struct {
	HitsTotal         prometheus.Counter
	Hits              *prometheus.CounterVec
	Times             *prometheus.HistogramVec
	PVZCreated        prometheus.Counter
	ReceptionsCreated prometheus.Counter
	ProductsAdded     prometheus.Counter
}

// Create metrics with address and name
func CreateMetrics(address, name string) (Metrics, error) {
	var metr PrometheusMetrics

	metr.HitsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: name + "_hits_total",
		Help: "Total number of HTTP requests",
	})
	if err := prometheus.Register(metr.HitsTotal); err != nil {
		return nil, err
	}

	metr.Hits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name + "_hits",
			Help: "Number of HTTP requests partitioned by status code, method and path",
		},
		[]string{"status", "method", "path"},
	)
	if err := prometheus.Register(metr.Hits); err != nil {
		return nil, err
	}

	metr.Times = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: name + "_response_time_seconds",
			Help: "Response time in seconds partitioned by status code, method and path",
		},
		[]string{"status", "method", "path"},
	)
	if err := prometheus.Register(metr.Times); err != nil {
		return nil, err
	}

	metr.PVZCreated = prometheus.NewCounter(prometheus.CounterOpts{
		Name: name + "_pvz_created_total",
		Help: "Total number of PVZ created",
	})
	if err := prometheus.Register(metr.PVZCreated); err != nil {
		return nil, err
	}

	metr.ReceptionsCreated = prometheus.NewCounter(prometheus.CounterOpts{
		Name: name + "_receptions_created_total",
		Help: "Total number of receptions created",
	})
	if err := prometheus.Register(metr.ReceptionsCreated); err != nil {
		return nil, err
	}

	metr.ProductsAdded = prometheus.NewCounter(prometheus.CounterOpts{
		Name: name + "_products_added_total",
		Help: "Total number of products added",
	})
	if err := prometheus.Register(metr.ProductsAdded); err != nil {
		return nil, err
	}

	if err := prometheus.Register(collectors.NewBuildInfoCollector()); err != nil {
		return nil, err
	}

	go func() {
		e := echo.New()
		e.HideBanner = true
		e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
		if err := e.Start(":" + address); err != nil {
			log.Fatal(err)
		}
	}()

	return &metr, nil
}

// IncHits
func (metr *PrometheusMetrics) IncHits(status int, method, path string) {
	metr.HitsTotal.Inc()
	metr.Hits.WithLabelValues(strconv.Itoa(status), method, path).Inc()
}

// Observe response time
func (metr *PrometheusMetrics) ObserveResponseTime(status int, method, path string, observeTime float64) {
	metr.Times.WithLabelValues(strconv.Itoa(status), method, path).Observe(observeTime)
}

// Inc PVZ created
func (metr *PrometheusMetrics) IncPVZCreated() {
	metr.PVZCreated.Inc()
}

// Inc receptions created
func (metr *PrometheusMetrics) IncReceptionsCreated() {
	metr.ReceptionsCreated.Inc()
}

// Inc products added
func (metr *PrometheusMetrics) IncProductsAdded() {
	metr.ProductsAdded.Inc()
}
