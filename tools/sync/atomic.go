package sync

import "sync/atomic"

type Atomic[T any] struct {
	value atomic.Value
}

func NewAtomic[T any](initial T) *Atomic[T] {
	v := &Atomic[T]{}
	v.value.Store(initial)

	return v
}

func (v *Atomic[T]) Load() T {
	return v.value.Load().(T)
}

func (v *Atomic[T]) Store(newValue T) {
	v.value.Store(newValue)
}

func (v *Atomic[T]) Swap(newValue T) T {
	return v.value.Swap(newValue).(T)
}

func (v *Atomic[T]) CompareAndSwap(oldValue, newValue T) bool {
	return v.value.CompareAndSwap(oldValue, newValue)
}