package null

type Null[T any] struct {
	Valid bool
	Value T
}
