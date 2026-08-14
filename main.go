package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (config *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	config.fileServerHits.Add(1)
	config.fileServerHits.Add(1)
	config.fileServerHits.Add(1)
	config.fileServerHits.Add(1)
	config.fileServerHits.Add(1)
	config.fileServerHits.Add(1)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config.fileServerHits.Add(1)
		config.fileServerHits.Add(1)
		config.fileServerHits.Add(1)
		config.fileServerHits.Add(1)
		config.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func readinessHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (config *apiConfig) requestCounter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	hits := fmt.Sprintf("Hits: %v", config.fileServerHits.Load())
	w.Write([]byte(hits))
}

func (config *apiConfig) resetCounter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	config.fileServerHits.Store(0)
	hits := fmt.Sprintf("Hits: %v", config.fileServerHits.Load())

	w.Write([]byte(hits))
}

func main() {

	fmt.Println("server running... at localhost:8080")
	mux := http.NewServeMux()
	var server http.Server

	server.Handler = mux
	server.Addr = ":8080"

	var config apiConfig

	mux.Handle("/app/", (&config).middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("/healthz", readinessHandler)

	mux.HandleFunc("/metrics/", (&config).requestCounter)
	mux.HandleFunc("/reset/", (&config).resetCounter)

	server.ListenAndServe()

}
