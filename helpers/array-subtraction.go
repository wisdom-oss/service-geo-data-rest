package helpers

import (
	"cmp"
	"slices"
)

func SubtractArrays[E cmp.Ordered](minuend []E, subtrahend []E) (difference []E) {
	slices.Sort(minuend)
	slices.Sort(subtrahend)
	for _, value := range subtrahend {
		idx, found := slices.BinarySearch(minuend, value)
		if found {
			minuend = slices.Delete(minuend, idx, idx+1)
		}
	}
	return minuend
}
