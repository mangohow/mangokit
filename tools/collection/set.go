package collection

type set[V comparable] struct {
	set map[V]struct{}
}

func NewSet[V comparable]() Set[V] {
	return &set[V]{set: make(map[V]struct{})}
}

func NewSetFromSlice[V comparable](s []V) Set[V] {
	st := &set[V]{set: make(map[V]struct{}, len(s))}
	for _, v := range s {
		st.Add(v)
	}

	return st
}

func (s *set[V]) Add(v V) {
	s.set[v] = struct{}{}
}

func (s *set[V]) Has(v V) bool {
	_, ok := s.set[v]
	return ok
}

func (s *set[V]) Delete(v V) {
	delete(s.set, v)
}

func (s *set[V]) Values() []V {
	return Keys(s.set)
}

func (s *set[V]) Any(vs []V) bool {
	for _, v := range vs {
		if s.Has(v) {
			return true
		}
	}

	return false
}

func (s *set[V]) Every(vs []V) bool {
	for _, v := range vs {
		if !s.Has(v) {
			return false
		}
	}

	return true
}

func (s *set[V]) ForEach(f func(v V)) {
	for k := range s.set {
		f(k)
	}
}

func (s *set[V]) ForEachP(f func(v *V)) {
	for k := range s.set {
		f(&k)
	}
}

func (s *set[V]) Len() int {
	return len(s.set)
}
