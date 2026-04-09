// Package packs calculates the optimal set of packs to fulfil an order.
//
// Rules (in priority order):
//  1. Only whole packs can be sent — packs cannot be broken open.
//  2. Send the fewest items possible to cover the order (primary).
//  3. Within the same item total, send the fewest packs possible (secondary).
package packs

import "errors"

// DefaultSizes is the standard set of pack sizes, ordered largest-first.
var DefaultSizes = []int{5000, 2000, 1000, 500, 250}

// Pack represents a quantity of a single pack size.
type Pack struct {
	Size  int `json:"size"`
	Count int `json:"count"`
}

// Result holds the full breakdown for an order.
type Result struct {
	Order      int    `json:"order"`
	Packs      []Pack `json:"packs"`
	TotalItems int    `json:"total_items"`
	TotalPacks int    `json:"total_packs"`
}

// Calculate returns the optimal packs needed to fulfil the given order using
// the provided pack sizes (must be sorted largest-first and all > 0).
//
// Algorithm:
//  1. Find the smallest multiple of the smallest pack size that covers the
//     order — this is the minimum items we can legally ship (Rule 2).
//  2. Greedily fill that target with the largest available packs first
//     to minimise pack count (Rule 3).
func Calculate(order int, sizes []int) (Result, error) {
	if order <= 0 {
		return Result{}, errors.New("order must be greater than zero")
	}
	if len(sizes) == 0 {
		return Result{}, errors.New("at least one pack size is required")
	}

	smallest := sizes[len(sizes)-1]

	// Rule 2: round up to the nearest multiple of the smallest pack size.
	rem := order % smallest
	target := order
	if rem != 0 {
		target = order + (smallest - rem)
	}

	// Rule 3: greedy fill from largest to smallest.
	var packs []Pack
	remaining := target
	for _, size := range sizes {
		if count := remaining / size; count > 0 {
			packs = append(packs, Pack{Size: size, Count: count})
			remaining -= count * size
		}
	}

	totalItems := 0
	totalPacks := 0
	for _, p := range packs {
		totalItems += p.Size * p.Count
		totalPacks += p.Count
	}

	return Result{
		Order:      order,
		Packs:      packs,
		TotalItems: totalItems,
		TotalPacks: totalPacks,
	}, nil
}
