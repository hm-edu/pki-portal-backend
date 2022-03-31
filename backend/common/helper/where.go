package helper

func Where[T any](a []T, f func(T) bool) []T {
	n := []T{}
	for _, e := range a {
		if f(e) {
			n = append(n, e)
		}
	}
	return n
}
