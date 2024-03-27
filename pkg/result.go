package pkg

type Result[T any] struct {
	V   T
	Err error
}
