package inline

import "fmt"

func If[T any](val bool, t T, f T) T {
	if val {
		return t
	}
	return f
}

func IfErrorOrNil(val bool, msg string, args ...interface{}) error {
	if val {
		return fmt.Errorf(msg, args...)
	}
	return nil
}