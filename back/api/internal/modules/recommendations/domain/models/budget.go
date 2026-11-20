package models

import "sync/atomic"

// Budget caps the requests a refresh makes to the catalog source.
type Budget struct {
	left atomic.Int64
}

func NewBudget(requests int) *Budget {
	b := &Budget{}
	b.left.Store(int64(requests))
	return b
}

// Take spends n requests and reports whether there were enough.
func (b *Budget) Take(n int) bool {
	for {
		left := b.left.Load()
		if left < int64(n) {
			return false
		}
		if b.left.CompareAndSwap(left, left-int64(n)) {
			return true
		}
	}
}

func (b *Budget) Left() int {
	return int(b.left.Load())
}
