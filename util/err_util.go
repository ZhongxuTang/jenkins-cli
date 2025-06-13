package util

func Must(e error) {
	if e != nil {
		panic(e)
	}
}
func MustAny[A any](a A, e error) A {
	if e != nil {
		panic(e)
	}
	return a
}
