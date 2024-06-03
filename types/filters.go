package types

type FilterType string

const (
	FilterWithin   FilterType = "within"
	FilterContains            = "contains"
	FilterOverlaps            = "overlaps"
)
