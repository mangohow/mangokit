package stream

import (
	"sort"
)

type pipelineStream[T any] struct {
	initial    []T
	progresses []func(T) (T, bool)
}

func newPipelineStream[T any](s []T) Stream[T] {
	return &pipelineStream[T]{
		initial: s,
	}
}

func (p *pipelineStream[T]) Filter(filter func(T) bool) Stream[T] {
	p.progresses = append(p.progresses, func(v T) (T, bool) {
		if filter(v) {
			return v, true
		}

		var zero T
		return zero, false
	})

	return p
}

func (p *pipelineStream[T]) Sorted(f func(T, T) bool) Stream[T] {
	p.initial = p.execute()
	p.progresses = make([]func(T) (T, bool), 0)
	sort.Slice(p.initial, func(i, j int) bool {
		return f(p.initial[i], p.initial[j])
	})

	return p
}

func (p *pipelineStream[T]) Peek(f func(T)) Stream[T] {
	p.progresses = append(p.progresses, func(v T) (T, bool) {
		f(v)
		return v, true
	})

	return p
}

func (p *pipelineStream[T]) Limit(i int) Stream[T] {
	index := 0
	p.progresses = append(p.progresses, func(v T) (T, bool) {
		if index < i {
			index++
			return v, true
		}

		var zero T
		return zero, false
	})

	return p
}

func (p *pipelineStream[T]) Skip(i int) Stream[T] {
	index := 0
	p.progresses = append(p.progresses, func(v T) (T, bool) {
		if index == i {
			var zero T
			return zero, false
		}

		return v, true
	})

	return p
}

func (p *pipelineStream[T]) execute() []T {
	size := 16
	if len(p.initial) < size {
		size = len(p.initial)
	}
	res := make([]T, 0, size)

	for i := 0; i < len(p.initial); i++ {
		if v, ok := p.executeOne(p.initial[i]); ok {
			res = append(res, v)
		}
	}

	return res
}

func (p *pipelineStream[T]) executeOne(e T) (T, bool) {
	var (
		v  T
		ok bool
	)

	for _, fn := range p.progresses {
		v, ok = fn(e)
		if !ok {
			var zero T
			return zero, false
		}
	}

	return v, true
}

func (p *pipelineStream[T]) ForEach(f func(T)) {
	for i := 0; i < len(p.initial); i++ {
		if v, ok := p.executeOne(p.initial[i]); ok {
			f(v)
		}
	}
}

func (p *pipelineStream[T]) Reduce(e T, f func(T, T) T) T {
	r := e
	for i := 0; i < len(p.initial); i++ {
		if v, ok := p.executeOne(p.initial[i]); ok {
			r = f(r, v)
		}
	}

	return r
}

func (p *pipelineStream[T]) CollectSlice() []T {
	return p.execute()
}

func (p *pipelineStream[T]) CollectStringMap(f func(T) string) map[string]T {
	res := make(map[string]T)
	for i := 0; i < len(p.initial); i++ {
		if v, ok := p.executeOne(p.initial[i]); ok {
			res[f(v)] = v
		}
	}

	return res
}

func (p *pipelineStream[T]) CollectIntMap(f func(T) int) map[int]T {
	res := make(map[int]T)
	for i := 0; i < len(p.initial); i++ {
		if v, ok := p.executeOne(p.initial[i]); ok {
			res[f(v)] = v
		}
	}

	return res
}

func (p *pipelineStream[T]) CollectMap(f func(T) any) map[any]T {
	res := make(map[any]T)
	for i := 0; i < len(p.initial); i++ {
		if v, ok := p.executeOne(p.initial[i]); ok {
			res[f(v)] = v
		}
	}

	return res
}

func (p *pipelineStream[T]) Collect(f func(T)) {
	for i := 0; i < len(p.initial); i++ {
		if v, ok := p.executeOne(p.initial[i]); ok {
			f(v)
		}
	}
}

func (p *pipelineStream[T]) Max(less func(T, T) bool) (T, bool) {
	if len(p.initial) == 0 {
		var zero T
		return zero, false
	}

	var (
		mx   T
		flag bool
	)
	for i := 0; i < len(p.initial); i++ {
		v, ok := p.executeOne(p.initial[i])
		if !ok {
			continue
		}

		if !flag {
			flag = true
			mx = v
			continue
		}

		if less(mx, v) {
			mx = v
		}
	}

	if flag {
		return mx, true
	}

	var zero T
	return zero, false
}

func (p *pipelineStream[T]) Min(less func(T, T) bool) (T, bool) {
	if len(p.initial) == 0 {
		var zero T
		return zero, false
	}

	var (
		mn   T
		flag bool
	)
	for i := 0; i < len(p.initial); i++ {
		v, ok := p.executeOne(p.initial[i])
		if !ok {
			continue
		}

		if !flag {
			flag = true
			mn = v
			continue
		}

		if less(v, mn) {
			mn = v
		}
	}

	if flag {
		return mn, true
	}

	var zero T
	return zero, false
}

func (p *pipelineStream[T]) Count() int {
	res := 0
	for i := 0; i < len(p.initial); i++ {
		if _, ok := p.executeOne(p.initial[i]); ok {
			res++
		}
	}

	return res
}

func (p *pipelineStream[T]) AnyMatch(f func(T) bool) bool {
	for i := 0; i < len(p.initial); i++ {
		if v, ok := p.executeOne(p.initial[i]); ok {
			if f(v) {
				return true
			}
		}
	}

	return false
}

func (p *pipelineStream[T]) AllMatch(f func(T) bool) bool {
	for i := 0; i < len(p.initial); i++ {
		if v, ok := p.executeOne(p.initial[i]); ok {
			if !f(v) {
				return false
			}
		}
	}

	return true
}

func (p *pipelineStream[T]) NoneMatch(f func(T) bool) bool {
	for i := 0; i < len(p.initial); i++ {
		if v, ok := p.executeOne(p.initial[i]); ok {
			if f(v) {
				return false
			}
		}
	}

	return true
}

func (p *pipelineStream[T]) First() (T, bool) {
	for i := 0; i < len(p.initial); i++ {
		if v, ok := p.executeOne(p.initial[i]); ok {
			return v, true
		}
	}

	var zero T
	return zero, false
}

func (p *pipelineStream[T]) Last() (T, bool) {
	for i := len(p.initial) - 1; i >= 0; i-- {
		if v, ok := p.executeOne(p.initial[i]); ok {
			return v, true
		}
	}

	var zero T
	return zero, false
}
