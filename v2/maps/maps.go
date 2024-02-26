package maps

import (
	"fmt"
	"strings"
)

type Map = map[string]interface{}
type ListOfMap = []Map

func JSON(args ...interface{}) map[string]interface{}{
	return Of[string, interface{}](args...)
}

func Of[K comparable, V any](args ...interface{}) map[K]V {

	if len(args)%2 != 0 {
		panic("some count must be even")
	}

	var k K
	var v V
	result := map[K]V{}

	for i, arg := range args {
		if i%2 == 1 {
			v = arg.(V)
			result[k] = v
		} else {
			k = arg.(K)
		}
	}

	return result

}

func New(args ...interface{}) map[interface{}]interface{} {

	if len(args)%2 != 0 {
		panic("some count must be even")
	}

	var k interface{}
	result := map[interface{}]interface{}{}

	for i, arg := range args {
		if i%2 == 0 {
			result[k] = arg
		} else {
			k = arg
		}
	}

	return result

}
func ToUrlQuery(m map[string]interface{}, replacers ...func(string, interface{}) interface{}) string {
	var values []string
	var replacer func(string, interface{}) interface{}

	if len(replacers) > 0 {
		replacer = replacers[0]
	}

	for k, v := range m {

		value := v

		if replacer != nil {
			value = replacer(k, v)
		}

		values = append(values,fmt.Sprintf("%v=%v", k, value))
	}
	return strings.Join(values, "&")
}
