package stream

type Stream[T any] interface {
	Filter(func(T) bool) Stream[T]
	Sorted(func(T, T) bool) Stream[T]
	Peek(func(T)) Stream[T]
	Limit(int) Stream[T]
	Skip(int) Stream[T]
	ForEach(func(T))
	Reduce(T, func(T, T) T) T
	CollectSlice() []T
	CollectStringMap(func(T) string) map[string]T
	CollectIntMap(func(T) int) map[int]T
	CollectMap(func(T) any) map[any]T
	Collect(func(T))
	Max(func(T, T) bool) (T, bool)
	Min(func(T, T) bool) (T, bool)
	Count() int
	AnyMatch(func(T) bool) bool
	AllMatch(func(T) bool) bool
	NoneMatch(func(T) bool) bool
	First() (T, bool)
	Last() (T, bool)
}
