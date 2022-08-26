package handlers

type Resolution string

const STATE Resolution = "state"
const DISTRICT Resolution = "district"
const ADMINISTRATION Resolution = "administration"
const MUNICIPAL Resolution = "municipal"

var shapeKeyLength = map[Resolution]int{
	STATE:          2,
	DISTRICT:       5,
	ADMINISTRATION: 9,
	MUNICIPAL:      12,
}

var invertedResolutionMap = map[string]Resolution{
	"state":          STATE,
	"district":       DISTRICT,
	"administration": ADMINISTRATION,
	"municipal":      MUNICIPAL,
}

func getShapeKeyLength(r string) int {
	resolution, resolutionAvailable := invertedResolutionMap[r]
	if !resolutionAvailable {
		return 0
	}
	return shapeKeyLength[resolution]
}
