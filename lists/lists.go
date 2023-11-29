package lists

import (
	_ "fmt"
	"reflect"
)

// Index returns the first index of the target interface{} `t`, or
// -1 if no match is found.
func Index(vs interface{}, t interface{}) int {

	ss := reflect.ValueOf(vs)
	s := reflect.Indirect(ss)

	for i := 0; i < s.Len(); i++ {
		it := s.Index(i)
		if it.Interface() == t {
			return i
		}
	}
	return -1
}

// Include returns `true` if the target interface{} t is in the
// slice.
func Include(vs interface{}, t interface{}) bool {
	return Index(vs, t) >= 0
}

// Any returns `true` if one of the interface{}s in the slice
// satisfies the predicate `f`.
func Any(vs interface{}, f func(interface{}) bool) bool {

	ss := reflect.ValueOf(vs)
	s := reflect.Indirect(ss)

	for i := 0; i < s.Len(); i++ {
		it := s.Index(i)
		if f(it.Interface()) {
			return true
		}
	}
	return false
}

// All returns `true` if all of the interface{}s in the slice
// satisfy the predicate `f`.
func All(vs interface{}, f func(interface{}) bool) bool {

	ss := reflect.ValueOf(vs)
	s := reflect.Indirect(ss)

	for i := 0; i < s.Len(); i++ {
		it := s.Index(i)
		if !f(it.Interface()) {
			return false
		}
	}
	return true
}

// Filter returns a new slice containing all interface{}s in the
// slice that satisfy the predicate `f`.
func FindAll(vs interface{}, f func(interface{}) bool) []interface{} {
	return Filter(vs, f)
}

func Filter(vs interface{}, f func(interface{}) bool) []interface{} {

	vsf := make([]interface{}, 0)

	ss := reflect.ValueOf(vs)
	s := reflect.Indirect(ss)

	for i := 0; i < s.Len(); i++ {
		it := s.Index(i)
		if f(it.Interface()) {
			vsf = append(vsf, it.Interface())
		}
	}
	return vsf
}

func Count(vs interface{}, f func(interface{}) bool) int {
	filtered := Filter(vs, f)
	return len(filtered)
}

// Filter returns a new slice containing all interface{}s in the
// slice that satisfy the predicate `f`.
func Find(vs interface{}, f func(interface{}) bool) interface{} {

	ss := reflect.ValueOf(vs)
	s := reflect.Indirect(ss)

	for i := 0; i < s.Len(); i++ {
		it := s.Index(i)
		if f(it.Interface()) {
			return it.Interface()
		}
	}
	return nil
}

// Map returns a new slice containing the results of applying
// the function `f` to each interface{} in the original slice.
func Map(vs interface{}, f func(interface{}) interface{}) []interface{} {

	ss := reflect.ValueOf(vs)
	s := reflect.Indirect(ss)

	vsm := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		it := s.Index(i)
		vsm[i] = f(it.Interface())
	}
	return vsm
}

func MapToInterface(vs interface{}) []interface{} {
	return Map(vs, func(i interface{}) interface{} {
		return i
	})
}

func Sort(vs interface{}, f func(interface{}, interface{}) int) {

	ss := reflect.ValueOf(vs)
	s := reflect.Indirect(ss)
	swap := reflect.Swapper(vs)

	for i := s.Len(); i > 0; i-- {
		//The inner loop will first iterate through the full length
		//the next iteration will be through n-1
		// the next will be through n-2 and so on
		for j := 1; j < i; j++ {

			v1 := ss.Index(j - 1)
			v2 := ss.Index(j)

			if f(v1.Interface(), v2.Interface()) > 0 {
				swap(j-1, j)
			}
		}
	}
}

func ListParts(vs interface{}, size int) [][]interface{} {

	all := [][]interface{}{}
	list := []interface{}{}

	ss := reflect.ValueOf(vs)
	s := reflect.Indirect(ss)

	for i := 0; i < s.Len(); i++ {

		it := s.Index(i)
		list = append(list, it.Interface())

		if len(list) >= size {
			all = append(all, list)
			list = []interface{}{}
		}

	}

	if len(list) > 0 {
		all = append(all, list)
	}

	return all

}

func UniqueValues(vs interface{}, uniqueValueResolver func(data interface{}) interface{}) []interface{} {
	return RemoveDuplicates(vs, uniqueValueResolver)
}

func RemoveDuplicates(vs interface{}, uniqueValueResolver func(data interface{}) interface{}) []interface{} {

	result := []interface{}{}
	ss := reflect.ValueOf(vs)
	s := reflect.Indirect(ss)

	for i := 0; i < s.Len(); i++ {

		it := s.Index(i)

		any := Any(result, func(item interface{}) bool {
			return uniqueValueResolver(item) == uniqueValueResolver(it.Interface())
		})

		if !any {
			result = append(result, it.Interface())
		}
	}

	return result
}
