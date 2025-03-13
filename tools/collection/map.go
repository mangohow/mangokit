package collection

import (
	"sync"
	"unsafe"
)

type concurrentMap[K comparable, V any] struct {
	m map[K]V
	sync.RWMutex
}

func NewConcurrentMap[K comparable, V any]() ConcurrentMap[K, V] {
	return &concurrentMap[K, V]{
		m: make(map[K]V),
	}
}

func NewConcurrentMapFromMap[K comparable, V any](m map[K]V) ConcurrentMap[K, V] {
	return &concurrentMap[K, V]{m: m}
}

func (c *concurrentMap[K, V]) Set(k K, v V) {
	c.Lock()
	c.m[k] = v
	c.Unlock()
}

func (c *concurrentMap[K, V]) Get(k K) (V, bool) {
	c.RLock()
	defer c.RUnlock()
	v, ok := c.m[k]
	return v, ok
}

func (c *concurrentMap[K, V]) GetBatch(ks []K) []V {
	res := make([]V, 0, len(ks))
	c.RLock()
	defer c.RUnlock()
	for _, k := range ks {
		if v, ok := c.m[k]; ok {
			res = append(res, v)
		}
	}
	return res
}

func (c *concurrentMap[K, V]) Delete(k K) {
	c.Lock()
	delete(c.m, k)
	c.Unlock()
}

func (c *concurrentMap[K, V]) Has(k K) bool {
	c.RLock()
	defer c.RUnlock()
	_, ok := c.m[k]
	return ok
}

func (c *concurrentMap[K, V]) Keys() []K {
	c.RLock()
	defer c.RUnlock()

	return Keys(c.m)
}

func (c *concurrentMap[K, V]) KeysSet() Set[K] {
	c.RLock()
	defer c.RUnlock()
	res := &set[K]{
		set: make(map[K]struct{}, len(c.m)),
	}

	for k := range c.m {
		res.set[k] = struct{}{}
	}

	return res
}

func (c *concurrentMap[K, V]) Values() []V {
	c.RLock()
	defer c.RUnlock()
	return Values(c.m)
}

func (c *concurrentMap[K, V]) Merge(other ConcurrentMap[K, V]) {
	if plm, ok := other.(PLM[K, V]); ok {
		c.mergeOther(plm)
		return
	}

	m := other.ToMap()
	c.Lock()
	defer c.Unlock()
	for k, v := range m {
		c.m[k] = v
	}
}

func (c *concurrentMap[K, V]) mergeOther(plm PLM[K, V]) {
	// 同一个map
	ptr := uintptr(unsafe.Pointer(c))
	if ptr == plm.Pointer() {
		return
	}
	// 根据地址决定加锁顺序
	if ptr < plm.Pointer() {
		c.Lock()
		defer c.Unlock()
		plm.RWLock().RLock()
		defer plm.RWLock().RUnlock()
	} else {
		plm.RWLock().RLock()
		defer plm.RWLock().RUnlock()
		c.Lock()
		defer c.Unlock()
	}

	m := plm.InnerMap()
	for k, v := range m {
		c.m[k] = v
	}
}

func (c *concurrentMap[K, V]) Clone() ConcurrentMap[K, V] {
	c.RLock()
	defer c.RUnlock()
	res := &concurrentMap[K, V]{
		m: make(map[K]V, len(c.m)),
	}
	for k, v := range c.m {
		res.m[k] = v
	}
	return res
}

func (c *concurrentMap[K, V]) ToMap() map[K]V {
	c.RLock()
	defer c.RUnlock()
	res := make(map[K]V, len(c.m))
	for k, v := range c.m {
		res[k] = v
	}
	return res
}

func (c *concurrentMap[K, V]) MergeMap(m map[K]V) {
	c.Lock()
	defer c.Unlock()
	for k, v := range m {
		c.m[k] = v
	}
}

func (c *concurrentMap[K, V]) Len() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.m)
}

func (c *concurrentMap[K, V]) Pointer() uintptr {
	return uintptr(unsafe.Pointer(c))
}

func (c *concurrentMap[K, V]) RWLock() *sync.RWMutex {
	return &c.RWMutex
}

func (c *concurrentMap[K, V]) InnerMap() map[K]V {
	return c.m
}
