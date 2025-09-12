package collections

import (
	"sync"

	"github.com/mobilemindtech/go-utils/v2/optional"
)

type KeyVal[K comparable, V any] struct {
	Key K
	Val V
}

func NewKeyVal[K comparable, V any](k K, v V) *KeyVal[K, V] {
	return &KeyVal[K, V]{Key: k, Val: v}
}

type ConcurrentMap[K comparable, V any] struct {
	data map[K]V
	lock sync.RWMutex
}

func NewConcurrentMap[K comparable, V any]() *ConcurrentMap[K, V] {
	return &ConcurrentMap[K, V]{data: map[K]V{}}
}

func (this *ConcurrentMap[K, V]) Put(k K, v V) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.data[k] = v
}

func (this *ConcurrentMap[K, V]) HasKey(k K) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	_, ok := this.data[k]
	return ok
}

func (this *ConcurrentMap[K, V]) Remove(k K) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	if _, ok := this.data[k]; ok {
		delete(this.data, k)
		return true
	}
	return false
}

func (this *ConcurrentMap[K, V]) Len() int {
	this.lock.Lock()
	defer this.lock.Unlock()
	return len(this.data)
}

func (this *ConcurrentMap[K, V]) Get(k K) (V, bool) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if v, ok := this.data[k]; ok {
		return v, ok
	}
	var v V
	return v, false
}

func (this *ConcurrentMap[K, V]) Each(each func(K, V)) {
	this.lock.Lock()
	defer this.lock.Unlock()
	for k, v := range this.data {
		each(k, v)
	}
}

func (this *ConcurrentMap[K, V]) TryGet(k K) *optional.Optional[V] {
	this.lock.Lock()
	defer this.lock.Unlock()
	if v, ok := this.data[k]; ok {
		return optional.WithSome[V](v)
	}
	return optional.WithNone[V]()
}

func (this *ConcurrentMap[K, V]) First() *optional.Optional[*KeyVal[K, V]] {
	this.lock.Lock()
	defer this.lock.Unlock()
	for k, v := range this.data {
		return optional.WithSome[*KeyVal[K, V]](NewKeyVal[K, V](k, v))
	}
	return optional.WithNone[*KeyVal[K, V]]()
}

func (this *ConcurrentMap[K, V]) Del(kv KeyVal[K, V]) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	if _, ok := this.data[kv.Key]; ok {
		delete(this.data, kv.Key)
		return true
	}
	return false
}
