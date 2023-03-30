package neo4j_test

func Contains[T comparable](items []T, search T) bool {
	for _, item := range items {
		if item == search {
			return true
		}
	}
	return false
}
