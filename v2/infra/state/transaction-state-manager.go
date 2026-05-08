package state

import (
	"fmt"
	"ismelen/ermc/v2/domain"
	"os"
	"sync"
	"time"
)

type TransactionStateManager struct {
	transactions map[string]*domain.Transaction
}

var transactionManager *TransactionStateManager
var once sync.Once
var mu sync.RWMutex

func GetManager() *TransactionStateManager { 
	if transactionManager == nil {
		once.Do(func() {
			transactionManager =  &TransactionStateManager{
				transactions: make(map[string]*domain.Transaction),
			}
		})
	}
	return transactionManager
}

func (t *TransactionStateManager) StartTransaction(id, path string, transactionSize int64) {
	mu.Lock()
	t.transactions[id] = &domain.Transaction{
		Id: id,
		StartAt: time.Now(),
		Size: transactionSize,
		Path: path,
	}
	mu.Unlock()
	
	time.AfterFunc(90*time.Minute, func() {
		t.DeleteTransaction(id)
	})
}

func (t *TransactionStateManager) UpdateProgress(id string, processedSize int64) { syncFunc(func(){
	tran, ok := t.transactions[id]
	if(!ok) { return }

	tran.Current += processedSize
}) }

func (t *TransactionStateManager) CheckProgress(id string) (int64, error) {
	mu.RLock()
	tran, ok := t.transactions[id]
	mu.RUnlock()

	if(!ok) { return 0, fmt.Errorf("transaction doesn't exists") }

	if tran.Error != nil {
		defer t.DeleteTransaction(id)
		return 0, tran.Error
	}

	return (tran.Current / tran.Size) * 100, nil
}

func (t *TransactionStateManager) SetDone(id string) { syncFunc(func() {
	tran, ok := t.transactions[id]
	if(!ok) { return }

	tran.Done = true
}) }

func (t *TransactionStateManager) SetError(id string, err error) {syncFunc(func() {
	tran, ok := t.transactions[id]
	if(!ok) { return }

	tran.Error = err
}) }

func (t *TransactionStateManager) SetResultPath(id string, path string) { syncFunc(func() {
	tran, ok := t.transactions[id]
	if(!ok) { return }

	tran.ResultPath = path
}) }

func (t *TransactionStateManager) DeleteTransaction(id string) { syncFunc(func() {
	tran, ok := t.transactions[id]
	if(!ok) { return }

	os.RemoveAll(tran.Path)
	delete(t.transactions, id)
}) }

func syncFunc(f func()) {
	mu.Lock()
	f()
	mu.Unlock()
}