package helper

// Any checks if any item in the passed list fullfils the specificed requirements and returns either true or false.
func Any[T any](a []T, f func(T) bool) bool {
	for _, e := range a {
		if f(e) {
			return true
		}
	}
	return false
}
