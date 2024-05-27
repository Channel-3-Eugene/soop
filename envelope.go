package soop

type Envelope[T any] struct {
	Item    T
	Retries map[string]uint8
}

func NewEnvelope[T any](item T) Envelope[T] {
	return Envelope[T]{Item: item, Retries: make(map[string]uint8)}
}
