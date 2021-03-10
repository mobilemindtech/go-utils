package lists

import "reflect"


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

func ListParts(vs interface{}, size int) [][]interface{} {

  all := [][]interface{}{}
  list := []interface{}{}

  ss := reflect.ValueOf(vs)    
  s := reflect.Indirect(ss)

  for i := 0; i < s.Len(); i++ {

    it := s.Index(i)
    list = append(list, it.interface())


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