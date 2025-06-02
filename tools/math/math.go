package math

type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

func Max[T Ordered](m, n T) T {
	if m > n {
		return m
	}
	return n
}

func Min[T Ordered](m, n T) T {
	if m < n {
		return m
	}
	return n
}
