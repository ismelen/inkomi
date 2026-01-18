package SharedModels

import "sync"

type SyncList struct {
	mu     sync.Mutex
	Values []string
}

func (this *SyncList) Add(value string) {
	this.mu.Lock()
	this.Values = append(this.Values, value)
	this.mu.Unlock()
}
