package collection

func Map[T any, U any](collection []T, fn func(T) U) []U {
	result := make([]U, 0, len(collection))
	for i := range collection {
		result = append(result, fn(collection[i]))
	}

	return result
}

func MapP[T any, U any](collection []T, fn func(*T) U) []U {
	result := make([]U, 0, len(collection))
	for i := range collection {
		result = append(result, fn(&collection[i]))
	}

	return result
}

func Filter[T any](collection []T, fn func(T) bool) []T {
	result := make([]T, 0, len(collection))
	for i := range collection {
		if fn(collection[i]) {
			result = append(result, collection[i])
		}
	}

	return result
}

func FilterP[T any](collection []T, fn func(*T) bool) []T {
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

func Sum[T Addable](args ...T) (res T) {
	for _, v := range args {
		res += v
	}

	return res
}
