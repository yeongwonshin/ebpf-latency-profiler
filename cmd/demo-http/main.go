package main

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		latency := time.Duration(30+rand.IntN(250)) * time.Millisecond
		time.Sleep(latency)
		if rand.IntN(100) < 5 {
			http.Error(w, "simulated downstream error", http.StatusBadGateway)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/orders/")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"id": id, "latency_ms": strconv.Itoa(int(latency.Milliseconds()))})
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	log.Println("demo HTTP service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
