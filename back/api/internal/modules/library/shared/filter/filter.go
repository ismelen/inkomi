package filter

// IFilter is a generic chain of responsibility link. It returns whether the chain produced
// a result and that result.
type IFilter[T any, R any] interface {
	SetNext(filter IFilter[T, R])
	Filter(items T) (bool, R)
}

type Base[T any, R any] struct {
	next IFilter[T, R]
}

func (b *Base[T, R]) SetNext(f IFilter[T, R]) { b.next = f }

func (b *Base[T, R]) Next(items T) (bool, R) {
	if b.next != nil {
		return b.next.Filter(items)
	}
	var zero R
	return false, zero
}

func Use[T any, R any](filters ...IFilter[T, R]) IFilter[T, R] {
	for i := 0; i < len(filters)-1; i++ {
		filters[i].SetNext(filters[i+1])
	}
	return filters[0]
}
