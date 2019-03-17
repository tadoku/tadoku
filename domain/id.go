package domain

// ContainsID returns wether or not an ID can be found in the array
func ContainsID(a []uint64, x uint64) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}

	return false
}
