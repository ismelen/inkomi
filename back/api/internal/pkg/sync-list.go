package pkg

import "sync"

type SyncList[T any] struct {
	mu     sync.Mutex
	Values []T
}

func NewSyncList[T any]() *SyncList[T] {
	return &SyncList[T]{}
}

func (this *SyncList[T]) Add(value T) {
	this.mu.Lock()
	this.Values = append(this.Values, value)
	this.mu.Unlock()
}
