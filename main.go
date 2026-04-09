package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"packapi/packs"
)

//go:embed ui/index.html
var ui embed.FS

// errorResponse is returned for all 4xx/5xx responses.
type errorResponse struct {
	Error string `json:"error"`
}

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// handlePacks handles GET /packs?order=<n>
//
// Query parameters:
//
//	order  int  (required) — number of items the customer wants
//
// Example:
//
//	GET /packs?order=12001
func handlePacks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{"method not allowed"})
		return
	}

	raw := r.URL.Query().Get("order")
	if raw == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"missing required query param: order"})
		return
	}

	const maxOrder = 1_000_000_000

	order, err := strconv.Atoi(raw)
	if err != nil || order <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{"order must be a positive integer"})
		return
	}
	if order > maxOrder {
		writeJSON(w, http.StatusBadRequest, errorResponse{"order must not exceed 1,000,000,000"})
		return
	}

	result, err := packs.Calculate(order, packs.DefaultSizes)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/packs", handlePacks)
	mux.Handle("/ui/", http.FileServer(http.FS(ui)))
	mux.Handle("/", http.RedirectHandler("/ui/index.html", http.StatusMovedPermanently))

	addr := ":8080"
	fmt.Printf("Pack calculator listening on %s\n", addr)
	fmt.Printf("  UI:  http://localhost%s\n", addr)
	fmt.Printf("  API: http://localhost%s/packs?order=12001\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
