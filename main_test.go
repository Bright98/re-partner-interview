package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"packapi/packs"
)

func TestMain(m *testing.M) {
	sizes = append([]int(nil), packs.DefaultSizes...)
	m.Run()
}

func TestHandlePacks_Success(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantPacks  []packs.Pack
		wantTotal  int
	}{
		{
			name:       "order 1",
			query:      "order=1",
			wantStatus: http.StatusOK,
			wantPacks:  []packs.Pack{{Size: 250, Count: 1}},
			wantTotal:  250,
		},
		{
			name:       "order 251",
			query:      "order=251",
			wantStatus: http.StatusOK,
			wantPacks:  []packs.Pack{{Size: 500, Count: 1}},
			wantTotal:  500,
		},
		{
			name:       "order 501",
			query:      "order=501",
			wantStatus: http.StatusOK,
			wantPacks:  []packs.Pack{{Size: 500, Count: 1}, {Size: 250, Count: 1}},
			wantTotal:  750,
		},
		{
			name:       "order 12001",
			query:      "order=12001",
			wantStatus: http.StatusOK,
			wantPacks:  []packs.Pack{{Size: 5000, Count: 2}, {Size: 2000, Count: 1}, {Size: 250, Count: 1}},
			wantTotal:  12250,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/packs?"+tc.query, nil)
			w := httptest.NewRecorder()

			handlePacks(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("status: got %d, want %d", w.Code, tc.wantStatus)
			}
			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type: got %q, want application/json", ct)
			}

			var result packs.Result
			if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if len(result.Packs) != len(tc.wantPacks) {
				t.Fatalf("packs length: got %d, want %d", len(result.Packs), len(tc.wantPacks))
			}
			for i, p := range result.Packs {
				if p != tc.wantPacks[i] {
					t.Errorf("pack[%d]: got %+v, want %+v", i, p, tc.wantPacks[i])
				}
			}
			if result.TotalItems != tc.wantTotal {
				t.Errorf("total_items: got %d, want %d", result.TotalItems, tc.wantTotal)
			}
		})
	}
}

func TestHandlePacks_BadRequests(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		query      string
		wantStatus int
	}{
		{
			name:       "missing order param",
			method:     http.MethodGet,
			query:      "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "non-numeric order",
			method:     http.MethodGet,
			query:      "order=abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "zero order",
			method:     http.MethodGet,
			query:      "order=0",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "negative order",
			method:     http.MethodGet,
			query:      "order=-10",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wrong method",
			method:     http.MethodPost,
			query:      "order=100",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := "/packs"
			if tc.query != "" {
				url += "?" + tc.query
			}
			req := httptest.NewRequest(tc.method, url, nil)
			w := httptest.NewRecorder()

			handlePacks(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tc.wantStatus)
			}

			var errResp errorResponse
			if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
				t.Fatalf("decode error response: %v", err)
			}
			if errResp.Error == "" {
				t.Error("expected non-empty error message in response body")
			}
		})
	}
}
