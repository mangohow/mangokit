package collection

func Map[T any, U any](collection []T, fn func(T) U) []U {
	if len(collection) == 0 {
		return nil
	}

	result := make([]U, 0, len(collection))
	for i := range collection {
		result = append(result, fn(collection[i]))
	}

	return result
}

func MapP[T any, U any](collection []T, fn func(*T) U) []U {
	if len(collection) == 0 {
		return nil
	}

	result := make([]U, 0, len(collection))
	for i := range collection {
		result = append(result, fn(&collection[i]))
	}

	return result
}

func Filter[T any](collection []T, fn func(T) bool) []T {
	if len(collection) == 0 {
		return nil
	}

	result := make([]T, 0, len(collection))
	for i := range collection {
		if fn(collection[i]) {
			result = append(result, collection[i])
		}
	}

	return result
}

func FilterP[T any](collection []T, fn func(*T) bool) []T {
	if len(collection) == 0 {
		return nil
	}

	result := make([]T, 0, len(collection))
	for i := range collection {
		if fn(&collection[i]) {
			result = append(result, collection[i])
		}
	}

	return result
}

func ForEach[T any](collection []T, fn func(T) bool) {
	for i := range collection {
		if !fn(collection[i]) {
			return
		}
	}
}

func ForEachP[T any](collection []T, fn func(*T) bool) {
	for i := range collection {
		if !fn(&collection[i]) {
			return
		}
	}
}

// 在原切片上修改
func Delete[T any, S []T](ss S, index int) S {
	if len(ss) == 0 {
		return ss
	}
	ss = append(ss[:index], ss[index+1:]...)
	return ss
}

// 在原切片上修改
func DeleteFunc[T any, S []T](ss S, fn func(T) bool) S {
	if len(ss) == 0 {
		return ss
	}
	keepIndex := 0
	for i, v := range ss {
		if !fn(v) {
			if i != keepIndex {
				ss[keepIndex] = v
			}
			keepIndex++
		}
	}

	clear(ss[keepIndex:])
	ss = ss[:keepIndex]

	return ss
}

func Every[S ~[]T, T any](slice S, f func(T) bool) bool {
	for _, v := range slice {
		if !f(v) {
			return false
		}
	}
	return true
}

func Some[S ~[]T, T any](slice S, f func(T) bool) bool {
	for _, v := range slice {
		if f(v) {
			return true
		}
	}
	return false
}

func Reduce[S ~[]T, T, U any](slice S, initial U, f func(U, T) U) U {
	accumulator := initial
	for _, v := range slice {
		accumulator = f(accumulator, v)
	}
	return accumulator
}

type Addable interface {
	~int | ~int64 | ~int32 | ~int16 | ~int8 | ~uint | ~uint64 | ~uint32 | ~uint16 | ~uint8 | ~float32 | ~float64 | ~uintptr
}

type IntNumber interface {
	~int | ~int64 | ~int32 | ~int16 | ~int8 | ~uint | ~uint64 | ~uint32 | ~uint16 | ~uint8
}

func Sum[T Addable](args ...T) (res T) {
	for _, v := range args {
		res += v
	}

	return res
}

type stream[T any] struct {
	ss []T
}

// 原切片不能再使用
func StreamOf[T any](ss []T) *stream[T] {
	return &stream[T]{ss}
}

func (s *stream[T]) Filter(fn func(T) bool) *stream[T] {
	s.ss = Filter(s.ss, fn)
	return s
}

func (s *stream[T]) ForEach(fn func(T) bool) *stream[T] {
	ForEach(s.ss, fn)
	return s
}

func (s *stream[T]) Find(fn func(T) bool) *stream[T] {
	ol := len(s.ss)
	s.ss = DeleteFunc(s.ss, func(v T) bool {
		return !fn(v)
	})

	// 如果跟之前长度差太多，为了防止内存泄漏，申请一个新的切片
	if len(s.ss) > 256 && len(s.ss) < int(ol/4) {
		tmp := make([]T, len(s.ss))
		copy(tmp, s.ss)
		s.ss = tmp
	}

	return s
}

func (s *stream[T]) FindFirst(fn func(T) bool) *stream[T] {
	for _, v := range s.ss {
		if fn(v) {
			s.ss = []T{v}
			return s
		}
	}

	s.ss = nil
	return s
}

func (s *stream[T]) Collect() []T {
	return s.ss
}

func (s *stream[T]) Every(fn func(T) bool) bool {
	return Every(s.ss, fn)
}

func (s *stream[T]) Some(fn func(T) bool) bool {
	return Some(s.ss, fn)
}

func (s *stream[T]) First() (T, bool) {
	if len(s.ss) == 0 {
		var zero T
		return zero, false
	}
	return s.ss[0], true
}

func (s *stream[T]) Last() (T, bool) {
	if len(s.ss) == 0 {
		var zero T
		return zero, false
	}
	return s.ss[len(s.ss)-1], true
}

// 左闭右闭
func SliceRange[T IntNumber](start, end T) []T {
	if start > end {
		return nil
	}
	res := make([]T, 0, end-start+1)
	for ; start <= end; start++ {
		res = append(res, start)
	}

	return res
}

func SliceGen[T any](n int, fn func() T) []T {
	if n <= 0 {
		return nil
	}
	res := make([]T, n)
	for i := 0; i < n; i++ {
		res[i] = fn()
	}

	return res
}
