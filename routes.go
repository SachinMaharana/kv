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
		Name: "kv",
		Help: "Duration of HTTP requests.",
	}, []string{"path", "method", "code"})

	keys = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "total_keys_kv",
			Help: "Total keys in kv store",
		},
		[]string{"keys"},
	)
)

func init() {
	prometheus.Register(keys)
}

func (app *application) routes() http.Handler {
	router := mux.NewRouter()
	// reg := prometheus.NewRegistry()
	// keys.WithLabelValues("keys").Set(float64(app.getTotalKeys()))

	router.Use(app.prometheusMiddleware)
	router.HandleFunc("/healthcheck", app.healthcheckHandler)
	router.HandleFunc("/search", app.search)
	router.HandleFunc("/get/{key}", app.getKey)
	router.HandleFunc("/set", app.setKey).Methods("POST")
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
		keys.WithLabelValues("keys").Set(float64(app.db.Total()))

		statusCode := rw.statusCode
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path, r.Method, strconv.Itoa(statusCode)))

		timer.ObserveDuration()
	})
}
