package enums

type Resolution string

const (
	STATE          Resolution = "state"
	DISTRICT       Resolution = "district"
	ADMINISTRATION Resolution = "administration"
	MUNICIPALITY   Resolution = "municipal"
)

var resolutionKeyLengthMapping = map[Resolution]int{
	STATE:          2,
	DISTRICT:       5,
	ADMINISTRATION: 9,
	MUNICIPALITY:   12,
}

// GetKeyLength returns the key length for the given resolution.
func (r Resolution) GetKeyLength() int {
	return resolutionKeyLengthMapping[r]
}
