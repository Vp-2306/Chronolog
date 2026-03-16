package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Vp-2306/Chronolog/internal/benchmark"
	"github.com/Vp-2306/Chronolog/internal/metrics"
)

type Engine interface {
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
}

type Server struct {
	engine Engine
	port   string
}

func NewServer(engine Engine, port string) *Server {
	return &Server{
		engine: engine,
		port:   port,
	}
}

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func (s *Server) Start() error {

	http.HandleFunc("/put", s.handlePut)
	http.HandleFunc("/get", s.handleGet)
	http.HandleFunc("/delete", s.handleDelete)
	http.HandleFunc("/metrics", s.handleMetrics)
	http.HandleFunc("/benchmark", s.handleBenchmark)

	// catch all OPTIONS preflight requests
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	})

	fmt.Println("ChronoLog server started on port", s.port)
	fmt.Println("Dashboard: http://localhost:" + s.port)

	return http.ListenAndServe(":"+s.port, nil)
}

func (s *Server) handlePut(w http.ResponseWriter, r *http.Request) {

	setCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")

	if key == "" || value == "" {
		http.Error(w, "key and value required", http.StatusBadRequest)
		return
	}

	if err := s.engine.Put([]byte(key), []byte(value)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"key":    key,
		"value":  value,
	})
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {

	setCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")

	if key == "" {
		http.Error(w, "key required", http.StatusBadRequest)
		return
	}

	value, err := s.engine.Get([]byte(key))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"key":    key,
		"value":  string(value),
	})
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {

	setCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")

	if key == "" {
		http.Error(w, "key required", http.StatusBadRequest)
		return
	}

	if err := s.engine.Delete([]byte(key)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"key":    key,
	})
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {

	setCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics.Global.Snapshot())
}

func (s *Server) handleBenchmark(w http.ResponseWriter, r *http.Request) {

	setCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	numOps := 1000

	chronoWrite, err := benchmark.RunWriteBenchmark(s.engine.Put, numOps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	chronoRead, err := benchmark.RunReadBenchmark(s.engine.Get, numOps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mapWrite := benchmark.RunMapWriteBenchmark(numOps)
	mapRead := benchmark.RunMapReadBenchmark(numOps)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chronolog_write": map[string]interface{}{
			"avg_latency_us": chronoWrite.AvgLatencyUs,
			"min_latency_ns": chronoWrite.MinLatencyNs,
			"max_latency_ns": chronoWrite.MaxLatencyNs,
			"ops_per_second": chronoWrite.OpsPerSecond,
			"total_ops":      chronoWrite.OperationCount,
		},
		"chronolog_read": map[string]interface{}{
			"avg_latency_us": chronoRead.AvgLatencyUs,
			"min_latency_ns": chronoRead.MinLatencyNs,
			"max_latency_ns": chronoRead.MaxLatencyNs,
			"ops_per_second": chronoRead.OpsPerSecond,
			"total_ops":      chronoRead.OperationCount,
		},
		"map_write": map[string]interface{}{
			"avg_latency_us": mapWrite.AvgLatencyUs,
			"min_latency_ns": mapWrite.MinLatencyNs,
			"max_latency_ns": mapWrite.MaxLatencyNs,
			"ops_per_second": mapWrite.OpsPerSecond,
			"total_ops":      mapWrite.OperationCount,
		},
		"map_read": map[string]interface{}{
			"avg_latency_us": mapRead.AvgLatencyUs,
			"min_latency_ns": mapRead.MinLatencyNs,
			"max_latency_ns": mapRead.MaxLatencyNs,
			"ops_per_second": mapRead.OpsPerSecond,
			"total_ops":      mapRead.OperationCount,
		},
		"summary": map[string]interface{}{
			"chronolog_write_format": benchmark.FormatResult("ChronoLog Write", chronoWrite),
			"chronolog_read_format":  benchmark.FormatResult("ChronoLog Read", chronoRead),
			"map_write_format":       benchmark.FormatResult("Map Write", mapWrite),
			"map_read_format":        benchmark.FormatResult("Map Read", mapRead),
		},
	})
}