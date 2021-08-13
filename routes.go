package main

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "kv_http_duration_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"path", "method", "code"})

	keys = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "total_keys_kv",
			Help: "total keys in kv store",
		},
		[]string{"kv"},
	)
	totalRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of get requests.",
		},
		[]string{"path"},
	)
	responseStatus = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "response_status",
			Help: "Status of HTTP response",
		},
		[]string{"status", "path"},
	)
)

func (app *application) routes() http.Handler {
	router := mux.NewRouter()

	router.Use(app.prometheusMiddleware)
	router.HandleFunc("/healthcheck", app.healthcheckHandler).Methods("GET")
	router.HandleFunc("/get/{key}", app.getKey).Methods("GET")
	router.HandleFunc("/set", app.setKey).Methods("POST")
	router.HandleFunc("/search", app.search).Methods("GET")
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	router.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)
	return router
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (app *application) prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)
		statusCode := rw.statusCode

		responseStatus.WithLabelValues(strconv.Itoa(statusCode), path).Inc()
		// TODO: best place to put this?
		totalRequests.WithLabelValues(path).Inc()
		keys.WithLabelValues("redis").Set(float64(app.db.TotalKeys()))
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path, r.Method, strconv.Itoa(statusCode)))
		timer.ObserveDuration()
	})
}
