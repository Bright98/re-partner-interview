package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"

	"packapi/packs"
)

//go:embed ui/index.html
var ui embed.FS

// errorResponse is returned for all 4xx/5xx responses.
type errorResponse struct {
	Error string `json:"error"`
}

var (
	mu    sync.RWMutex
	sizes []int
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// handlePacks handles GET /packs?order=<n>
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

	mu.RLock()
	current := append([]int(nil), sizes...)
	mu.RUnlock()

	result, err := packs.Calculate(order, current)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// handlePackSizes handles GET/POST/DELETE /pack-sizes.
// Changes are in-memory only — the config file is never modified.
//
//	GET    — list current sizes
//	POST   — add a size    body: {"size": N}
//	DELETE — remove a size ?size=N
func handlePackSizes(w http.ResponseWriter, r *http.Request) {
	type response struct {
		PackSizes []int `json:"pack_sizes"`
	}

	switch r.Method {
	case http.MethodGet:
		mu.RLock()
		current := append([]int(nil), sizes...)
		mu.RUnlock()
		writeJSON(w, http.StatusOK, response{current})

	case http.MethodPost:
		var body struct {
			Size int `json:"size"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Size <= 0 {
			writeJSON(w, http.StatusBadRequest, errorResponse{"body must be JSON with a positive integer 'size' field"})
			return
		}

		mu.Lock()
		for _, s := range sizes {
			if s == body.Size {
				mu.Unlock()
				writeJSON(w, http.StatusConflict, errorResponse{fmt.Sprintf("pack size %d already exists", body.Size)})
				return
			}
		}
		sizes = append(sizes, body.Size)
		sort.Sort(sort.Reverse(sort.IntSlice(sizes)))
		current := append([]int(nil), sizes...)
		mu.Unlock()
		writeJSON(w, http.StatusCreated, response{current})

	case http.MethodDelete:
		size, err := strconv.Atoi(r.URL.Query().Get("size"))
		if err != nil || size <= 0 {
			writeJSON(w, http.StatusBadRequest, errorResponse{"size query param must be a positive integer"})
			return
		}

		mu.Lock()
		if len(sizes) == 1 {
			mu.Unlock()
			writeJSON(w, http.StatusBadRequest, errorResponse{"cannot remove the last pack size"})
			return
		}
		updated := sizes[:0]
		found := false
		for _, s := range sizes {
			if s == size {
				found = true
			} else {
				updated = append(updated, s)
			}
		}
		if !found {
			mu.Unlock()
			writeJSON(w, http.StatusNotFound, errorResponse{fmt.Sprintf("pack size %d not found", size)})
			return
		}
		sizes = updated
		current := append([]int(nil), sizes...)
		mu.Unlock()
		writeJSON(w, http.StatusOK, response{current})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{"method not allowed"})
	}
}

func main() {
	sizes = append([]int(nil), packs.DefaultSizes...)

	mux := http.NewServeMux()
	mux.HandleFunc("/packs", handlePacks)
	mux.HandleFunc("/pack-sizes", handlePackSizes)
	mux.Handle("/ui/", http.FileServer(http.FS(ui)))
	mux.Handle("/", http.RedirectHandler("/ui/index.html", http.StatusMovedPermanently))

	addr := ":8080"
	fmt.Printf("Pack calculator listening on %s\n", addr)
	fmt.Printf("  UI:  http://localhost%s\n", addr)
	fmt.Printf("  API: http://localhost%s/packs?order=12001\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
