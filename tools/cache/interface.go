package cache

type DBCache[K comparable, V any] interface {
	// load data from db
	Load() error
	Get(K) (V, bool)
	GetBatch([]K) []V
	GetAll() map[K]V
	Insert(V) error
	Update(V) error
	Delete(V) error
}

type SelectFunc[V any] func() ([]V, error)
type InsertFunc[V any] func(V) error
type UpdateFunc[V any] func(V) error
type DeleteFunc[V any] func(V) error
type KeyFunc[K comparable, V any] func(V) K
