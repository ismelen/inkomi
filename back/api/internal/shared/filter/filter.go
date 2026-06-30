package filter

// Filter is a generic chain of responsibility interface.
// It returns (passed bool, result R) where R is the result type.
type Filter[T any, R any] interface {
	SetNext(filter Filter[T, R])
	Filter(items T) (bool, R)
}

type Base[T any, R any] struct {
	next Filter[T, R]
}

func (b *Base[T, R]) SetNext(f Filter[T, R]) { b.next = f }

func (b *Base[T, R]) Next(items T) (bool, R) {
	if b.next != nil {
		return b.next.Filter(items)
	}
	var zero R
	return false, zero
}

func Use[T any, R any](filters ...Filter[T, R]) Filter[T, R] {
	for i := 0; i < len(filters)-1; i++ {
		filters[i].SetNext(filters[i+1])
	}
	return filters[0]
}
