package helper

// Where filters the passed list and returns a new slice. This can be also empty.
func Where[T any](a []T, f func(T) bool) []T {
	n := []T{}
	for _, e := range a {
		if f(e) {
			n = append(n, e)
		}
	}
	return n
}
