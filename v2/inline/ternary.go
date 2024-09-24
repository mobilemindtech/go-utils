package inline

func If[T any](val bool, t T, f T) T {
	if val {
		return t
	}
	return f
}
