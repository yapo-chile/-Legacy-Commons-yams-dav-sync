package infrastructure

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/loggers"
)

// Prometheus exposes application metrics, and profile runtime performance in
// /metrics path
type Prometheus struct {
	Server *http.Server
	// common  metrics for handlers
	requestDuration   *prometheus.HistogramVec
	requestCounterVec *prometheus.CounterVec
	requestSize       *prometheus.HistogramVec
	responseSize      *prometheus.HistogramVec
	// custom metrics
	processedImages  prometheus.Counter
	skippedImages    prometheus.Counter
	notFoundImages   prometheus.Counter
	sentImages       prometheus.Counter
	failedUploads    prometheus.Counter
	duplicatedImages prometheus.Counter
	recoveredImages  prometheus.Counter
	totalImages      prometheus.Gauge
	logger           loggers.Logger
}

// NewPrometheusExporter generate a new prometheus instance
func NewPrometheusExporter(port string) interfaces.MetricsExposer {
	// Initialize exposed metrics
	p := Prometheus{
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Histogram of latencies for HTTP requests.",
				Buckets: []float64{.05, 0.1, .25, .5, .75, 1, 2, 5, 20, 60},
			},
			[]string{"handler", "method"},
		),
		requestCounterVec: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_request_total",
				Help: "Counter of HTTP request to the endpoint.",
			},
			[]string{"handler"},
		),
		requestSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_size_bytes",
				Help:    "Histogram of request size for HTTP requests.",
				Buckets: prometheus.ExponentialBuckets(100, 10, 7),
			},
			[]string{"handler", "method"},
		),
		responseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "Histogram of response size for HTTP requests.",
				Buckets: prometheus.ExponentialBuckets(100, 10, 7),
			},
			[]string{"handler", "method"},
		),
		sentImages: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "yams_sent_images_total",
				Help: "Total sent images to yams",
			},
		),
		processedImages: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "yams_processed_images_total",
				Help: "Total processed images",
			},
		),
		skippedImages: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "yams_skipped_images_total",
				Help: "Total skipped images",
			},
		),
		notFoundImages: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "yams_not_found_images_total",
				Help: "Total not found in local storage",
			},
		),
		failedUploads: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "yams_failed_images_total",
				Help: "Total failed uploads",
			},
		),
		duplicatedImages: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "yams_duplicated_images_total",
				Help: "Total of images already in yams bucket",
			},
		),
		recoveredImages: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "yams_recovered_images_total",
				Help: "Total of failed images in previous upload and now they were uploaded correctly",
			},
		),
		totalImages: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "yams_images_total",
				Help: "Total of images to be sent to yams",
			},
		),
	}
	prometheus.MustRegister(p.requestSize)
	prometheus.MustRegister(p.requestDuration)
	prometheus.MustRegister(p.responseSize)
	prometheus.MustRegister(p.requestCounterVec)
	prometheus.MustRegister(p.sentImages)
	prometheus.MustRegister(p.processedImages)
	prometheus.MustRegister(p.skippedImages)
	prometheus.MustRegister(p.failedUploads)
	prometheus.MustRegister(p.notFoundImages)
	prometheus.MustRegister(p.duplicatedImages)
	prometheus.MustRegister(p.recoveredImages)
	prometheus.MustRegister(p.totalImages)

	p.expose(port)
	return &p
}

// InstrumentHandler wraps a HandlerFunc exposing metrics of request duration & size in prometheus
func (p Prometheus) InstrumentHandler(handlerName string, handler http.HandlerFunc) http.HandlerFunc {
	handler =
		// first metric: RequestDuration, this will be stored in requestDuration var
		promhttp.InstrumentHandlerDuration(
			p.requestDuration.MustCurryWith(prometheus.Labels{"handler": handlerName}),

			// Next metrics: Response size, this will be stored in responseSize var
			promhttp.InstrumentHandlerResponseSize(
				p.responseSize.MustCurryWith(prometheus.Labels{"handler": handlerName}),

				// Next metric: Request size, this will be stored in requestSize var
				promhttp.InstrumentHandlerRequestSize(
					p.requestSize.MustCurryWith(prometheus.Labels{"handler": handlerName}),

					// Next metric: Handler counter, this will be stored in requestCounterVec var
					promhttp.InstrumentHandlerCounter(
						p.requestCounterVec.MustCurryWith(prometheus.Labels{"handler": handlerName}),

						// Replace this handler to add new metrics
						handler,
					),
				),
			),
		)

	// return tracked handler
	return handler
}

// IncrementCounter increments a prometheus counter for a given metric
func (p *Prometheus) IncrementCounter(metric int) {
	switch metric {
	case domain.SentImages:
		p.sentImages.Inc()
	case domain.ProcessedImages:
		p.processedImages.Inc()
	case domain.SkippedImages:
		p.skippedImages.Inc()
	case domain.NotFoundImages:
		p.notFoundImages.Inc()
	case domain.FailedUploads:
		p.failedUploads.Inc()
	case domain.DuplicatedImages:
		p.duplicatedImages.Inc()
	case domain.RecoveredImages:
		p.recoveredImages.Inc()
	}
}

// SetGauge set a gauge for given metric
func (p *Prometheus) SetGauge(metric int, value float64) {
	switch metric {
	case domain.TotalImages:
		p.totalImages.Set(value)
	}
}

// expose starts prometheus exporter metrics server exposing metrics in "/metrics" path
func (p *Prometheus) expose(port string) {
	p.Server = &http.Server{Addr: ":" + port}
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := p.Server.ListenAndServe(); err != http.ErrServerClosed {
			p.logger.Error("Prometheus: %s", err)
		}
	}()

}

// Close closes prometheus server
func (p *Prometheus) Close() error {
	return p.Server.Close()
}
