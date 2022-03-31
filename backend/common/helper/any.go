package helper

func Any[T any](a []T, f func(T) bool) bool {
	for _, e := range a {
		if f(e) {
			return true
		}
	}
	return false
}
