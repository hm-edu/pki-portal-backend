package helper

// First gets the first item in the passed list that fullfils the specificed requirements or returns nil.
func First[T any](a []*T, f func(*T) bool) *T {
	for _, e := range a {
		if f(e) {
			return e
		}
	}
	return nil
}
