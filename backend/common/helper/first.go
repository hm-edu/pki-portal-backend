package helper

func First[T any](a []*T, f func(*T) bool) *T {
	for _, e := range a {
		if f(e) {
			return e
		}
	}
	return nil
}
