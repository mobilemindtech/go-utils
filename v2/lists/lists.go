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

func Empty[T any](vs []T, f func(T) bool) bool {
	return !Any[T](vs, f)
}

func Any[T any](vs []T, f func(T) bool) bool {
	for _, it := range vs {
		if f(it) {
			return true
		}
	}
	return false
}

// Filter returns a new slice containing all interface{}s in the
// slice that satisfy the predicate `f`.
func FindAll[T any](vs []T, f func(T) bool) []T {
	return Filter[T](vs, f)
}

func Filter[T any](vs []T, f func(T) bool) []T {
	vsf := []T{}
	for _, it := range vs {
		if f(it) {
			vsf = append(vsf, it)
		}
	}
	return vsf
}

// Filter returns a new slice containing all interface{}s in the
// slice that satisfy the predicate `f`.
func Find[T any](vs []T, f func(T) bool) T {
	var x T
	for _, it := range vs {
		if f(it) {
			return it
		}
	}
	return x
}

// Map returns a new slice containing the results of applying
// the function `f` to each interface{} in the original slice.
func Map[T any, R any](vs []T, f func(T) R) []R {
	vsf := []R{}
	for _, it := range vs {
		vsf = append(vsf, f(it))
	}
	return vsf
}

func Sort[T any](vs []T, f func(T, T) int) {

	l := len(vs)
	swap := reflect.Swapper(vs)

	for i := l; i > 0; i-- {
		//The inner loop will first iterate through the full length
		//the next iteration will be through n-1
		// the next will be through n-2 and so on
		for j := 1; j < i; j++ {

			v1 := vs[j-1]
			v2 := vs[j]

			if f(v1, v2) > 0 {
				swap(j-1, j)
			}
		}
	}
}

func FoldLeft[T any, Acc any](vs []T, initial Acc, fold func(Acc, T) Acc) Acc {
	nextAcc := initial
	for _, it := range vs {
		nextAcc = fold(nextAcc, it)
	}
	return nextAcc
}

func ListParts[T any](vs []T, size int) [][]T {

	all := [][]T{}
	list := []T{}

	for _, it := range vs {

		list = append(list, it)

		if len(list) >= size {
			all = append(all, list)
			list = []T{}
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

func Split[T any](vs []T, size int) [][]T {

	all := [][]T{}
	list := []T{}

	for _, it := range vs {

		list = append(list, it)

		if len(list) >= size {
			all = append(all, list)
			list = []T{}
		}

	}

	if len(list) > 0 {
		all = append(all, list)
	}

	return all

}

func Foreach[T any](vs []T, each func(T)) {
	for _, it := range vs {
		each(it)
	}
}

type CrudList[T any] struct {
	SaveList   []T
	RemoveList []T
	UpdateList []T
}

func EmptyCrudList[T any]() *CrudList[T] {
	return &CrudList[T]{SaveList: []T{}, RemoveList: []T{}, UpdateList: []T{}}
}

func NewCrudList[T any](currentList []T, newList []T, comparator func(T, T) bool) *CrudList[T] {
	crud := EmptyCrudList[T]()

	findInCurrentList := func(newVal T) bool {
		return Any[T](currentList, func(x T) bool { return comparator(x, newVal) })
	}

	notInCurrentList := func(newVal T) bool {
		return !Any[T](currentList, func(x T) bool { return comparator(x, newVal) })
	}

	notInNewList := func(newVal T) bool {
		return !Any[T](newList, func(x T) bool { return comparator(x, newVal) })
	}

	crud.SaveList = FindAll(newList, notInCurrentList)
	crud.UpdateList = FindAll(newList, findInCurrentList)
	crud.RemoveList = FindAll(currentList, notInNewList)

	return crud
}
