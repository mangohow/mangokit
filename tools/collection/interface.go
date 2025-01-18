package collection

import "sync"

type ConcurrentMap[K comparable, V any] interface {
	Set(key K, v V)
	Get(key K) (V, bool)
	GetBatch(keys []K) []V
	Delete(key K)
	Has(key K) bool
	Keys() []K
	KeysSet() Set[K]
	Values() []V
	Merge(ConcurrentMap[K, V])
	Clone() ConcurrentMap[K, V]
	ToMap() map[K]V
	MergeMap(map[K]V)
	Len() int
}

type PLM[K comparable, V any] interface {
	// Pointer 获取地址
	Pointer() uintptr
	// RWLock 获取内部的RWMutex
	RWLock() *sync.RWMutex
	// InnerMap 获取内部map
	InnerMap() map[K]V
}

type Set[V comparable] interface {
	Add(v V)
	Has(v V) bool
	Delete(v V)
	Values() []V
	Any([]V) bool
	Every([]V) bool
	ForEach(func(v V))
	ForEachP(func(v *V))
	Len() int
}
