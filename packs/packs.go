// Package packs calculates the optimal set of packs to fulfil an order.
//
// Rules (in priority order):
//  1. Only whole packs can be sent — packs cannot be broken open.
//  2. Send the fewest items possible to cover the order (primary).
//  3. Within the same item total, send the fewest packs possible (secondary).
package packs

import (
	"errors"
	"math"
	"sort"
)

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

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// Calculate returns the optimal packs needed to fulfil the given order.
//
// Algorithm:
//
//	For each combination of the smaller pack counts (bounded by largest/gcd(largest, size),
//	beyond which counts are redundant), compute how many of the largest pack are needed to
//	cover the remainder. Track the combination with fewest total items, then fewest packs.
//
// This handles arbitrary pack sizes correctly, including cases where the greedy
// "round up to nearest multiple of smallest" approach fails.
func Calculate(order int, sizes []int) (Result, error) {
	if order <= 0 {
		return Result{}, errors.New("order must be greater than zero")
	}
	if len(sizes) == 0 {
		return Result{}, errors.New("at least one pack size is required")
	}

	sorted := append([]int(nil), sizes...)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))

	bestCounts := solve(order, sorted)

	var packList []Pack
	totalItems, totalPacks := 0, 0
	for i, s := range sorted {
		if bestCounts[i] > 0 {
			packList = append(packList, Pack{Size: s, Count: bestCounts[i]})
			totalItems += s * bestCounts[i]
			totalPacks += bestCounts[i]
		}
	}

	return Result{
		Order:      order,
		Packs:      packList,
		TotalItems: totalItems,
		TotalPacks: totalPacks,
	}, nil
}

// solve returns the optimal count for each size (parallel to sizes, largest-first).
func solve(order int, sizes []int) []int {
	largest := sizes[0]
	smaller := sizes[1:]

	// For each smaller size s, counts beyond largest/gcd(largest,s) are redundant:
	// that many packs of s equal an integer number of packs of 'largest', so they
	// can always be swapped without changing the total — or improving pack count.
	bounds := make([]int, len(smaller))
	for i, s := range smaller {
		bounds[i] = largest / gcd(largest, s)
	}

	bestTotal := math.MaxInt
	bestPacks := math.MaxInt
	bestCounts := make([]int, len(sizes))

	cur := make([]int, len(smaller)) // current counts for each smaller size
	for {
		partial, partialPacks := 0, 0
		for i, s := range smaller {
			partial += cur[i] * s
			partialPacks += cur[i]
		}

		largeCount := 0
		if partial < order {
			largeCount = (order - partial + largest - 1) / largest
		}
		total := largeCount*largest + partial
		numPacks := largeCount + partialPacks

		if total < bestTotal || (total == bestTotal && numPacks < bestPacks) {
			bestTotal = total
			bestPacks = numPacks
			bestCounts[0] = largeCount
			copy(bestCounts[1:], cur)
		}

		// Odometer-style increment across all smaller-size counters.
		i := len(cur) - 1
		for i >= 0 {
			cur[i]++
			if cur[i] < bounds[i] {
				break
			}
			cur[i] = 0
			i--
		}
		if i < 0 {
			break
		}
	}

	return bestCounts
}
