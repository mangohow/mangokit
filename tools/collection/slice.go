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

func ForEach[T any](collection []T, fn func(T)) {
	for i := range collection {
		fn(collection[i])
	}
}

func ForEachP[T any](collection []T, fn func(*T)) {
	for i := range collection {
		fn(&collection[i])
	}
}
