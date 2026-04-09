package packs_test

import (
	"reflect"
	"testing"

	"packapi/packs"
)

func TestCalculate(t *testing.T) {
	sizes := packs.DefaultSizes

	tests := []struct {
		name       string
		order      int
		wantPacks  []packs.Pack
		wantTotal  int
		wantNPacks int
	}{
		{
			name:       "order 1 → 1x250 (smallest possible pack)",
			order:      1,
			wantPacks:  []packs.Pack{{250, 1}},
			wantTotal:  250,
			wantNPacks: 1,
		},
		{
			name:       "order 250 → exact 1x250",
			order:      250,
			wantPacks:  []packs.Pack{{250, 1}},
			wantTotal:  250,
			wantNPacks: 1,
		},
		{
			name:       "order 251 → 1x500 (not 2x250, fewer packs)",
			order:      251,
			wantPacks:  []packs.Pack{{500, 1}},
			wantTotal:  500,
			wantNPacks: 1,
		},
		{
			name:       "order 500 → exact 1x500",
			order:      500,
			wantPacks:  []packs.Pack{{500, 1}},
			wantTotal:  500,
			wantNPacks: 1,
		},
		{
			name:       "order 501 → 1x500 + 1x250 (not 1x1000, fewer items)",
			order:      501,
			wantPacks:  []packs.Pack{{500, 1}, {250, 1}},
			wantTotal:  750,
			wantNPacks: 2,
		},
		{
			name:       "order 750 → 1x500 + 1x250",
			order:      750,
			wantPacks:  []packs.Pack{{500, 1}, {250, 1}},
			wantTotal:  750,
			wantNPacks: 2,
		},
		{
			name:       "order 1000 → exact 1x1000",
			order:      1000,
			wantPacks:  []packs.Pack{{1000, 1}},
			wantTotal:  1000,
			wantNPacks: 1,
		},
		{
			name:       "order 1001 → 1x1000 + 1x250",
			order:      1001,
			wantPacks:  []packs.Pack{{1000, 1}, {250, 1}},
			wantTotal:  1250,
			wantNPacks: 2,
		},
		{
			name:       "order 2000 → exact 1x2000",
			order:      2000,
			wantPacks:  []packs.Pack{{2000, 1}},
			wantTotal:  2000,
			wantNPacks: 1,
		},
		{
			name:       "order 5000 → exact 1x5000",
			order:      5000,
			wantPacks:  []packs.Pack{{5000, 1}},
			wantTotal:  5000,
			wantNPacks: 1,
		},
		{
			name:       "order 5001 → 1x5000 + 1x250",
			order:      5001,
			wantPacks:  []packs.Pack{{5000, 1}, {250, 1}},
			wantTotal:  5250,
			wantNPacks: 2,
		},
		{
			// spec example: 2x5000, 1x2000, 1x250 = 12250 items
			name:       "order 12001 → 2x5000 + 1x2000 + 1x250",
			order:      12001,
			wantPacks:  []packs.Pack{{5000, 2}, {2000, 1}, {250, 1}},
			wantTotal:  12250,
			wantNPacks: 4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := packs.Calculate(tc.order, sizes)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got.Packs, tc.wantPacks) {
				t.Errorf("packs: got %v, want %v", got.Packs, tc.wantPacks)
			}
			if got.TotalItems != tc.wantTotal {
				t.Errorf("total items: got %d, want %d", got.TotalItems, tc.wantTotal)
			}
			if got.TotalPacks != tc.wantNPacks {
				t.Errorf("total packs: got %d, want %d", got.TotalPacks, tc.wantNPacks)
			}
		})
	}
}

func TestCalculate_ArbitrarySizes(t *testing.T) {
	// Pack sizes with no common divisor — greedy "round to smallest" fails here.
	// 500,000 = 23×2 + 31×7 + 53×9429 (exact, no overshoot needed)
	got, err := packs.Calculate(500_000, []int{23, 31, 53})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TotalItems != 500_000 {
		t.Errorf("total items: got %d, want 500000", got.TotalItems)
	}
	want := []packs.Pack{{Size: 53, Count: 9429}, {Size: 31, Count: 7}, {Size: 23, Count: 2}}
	if !reflect.DeepEqual(got.Packs, want) {
		t.Errorf("packs: got %v, want %v", got.Packs, want)
	}
}

func TestCalculate_Errors(t *testing.T) {
	t.Run("zero order", func(t *testing.T) {
		_, err := packs.Calculate(0, packs.DefaultSizes)
		if err == nil {
			t.Error("expected error for order=0")
		}
	})

	t.Run("negative order", func(t *testing.T) {
		_, err := packs.Calculate(-5, packs.DefaultSizes)
		if err == nil {
			t.Error("expected error for negative order")
		}
	})

	t.Run("no pack sizes", func(t *testing.T) {
		_, err := packs.Calculate(100, nil)
		if err == nil {
			t.Error("expected error for empty sizes")
		}
	})
}
