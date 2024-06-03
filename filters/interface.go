package filters

type Filter interface {
	// BuildQueryPart builds the part which is used in the 'WHERE' part of the SQL
	// query
	BuildQueryPart(layer string, keys ...string) (string, error)
}
