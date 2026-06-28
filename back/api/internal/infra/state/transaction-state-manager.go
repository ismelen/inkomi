package state

import (
	"fmt"
	"ismelen/inkomi/internal/domain"
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
			transactionManager = &TransactionStateManager{
				transactions: make(map[string]*domain.Transaction),
			}
		})
	}
	return transactionManager
}

func (t *TransactionStateManager) StartTransaction(id, path string, pages int) {
	mu.Lock()
	t.transactions[id] = &domain.Transaction{
		Id:      id,
		StartAt: time.Now(),
		Pages:   pages,
		Path:    path,
	}
	mu.Unlock()

	time.AfterFunc(90*time.Minute, func() {
		t.DeleteTransaction(id)
	})
}

func (t *TransactionStateManager) UpdateProgress(id string, processedPages int) bool {
	mu.Lock()
	defer mu.Unlock()

	tran, ok := t.transactions[id]
	if !ok {
		return false
	}
	tran.Current += processedPages

	return !tran.Canceled
}

func (t *TransactionStateManager) CheckProgress(id string) (int, error) {
	mu.RLock()
	tran, ok := t.transactions[id]
	mu.RUnlock()

	if !ok {
		return 0, fmt.Errorf("transaction doesn't exists")
	}

	if tran.Error != nil {
		return 0, tran.Error
	}

	return tran.Current * 100 / tran.Pages, nil
}

func (t *TransactionStateManager) SetDone(id string) {
	syncFunc(func() {
		tran, ok := t.transactions[id]
		if !ok {
			return
		}

		tran.Done = true
		tran.Current = tran.Pages
	})
}

func (t *TransactionStateManager) SetError(id string, err error) {
	syncFunc(func() {
		tran, ok := t.transactions[id]
		if !ok {
			return
		}

		tran.Error = err
	})
}

func (t *TransactionStateManager) Cancel(id string) {
	syncFunc(func() {
		tran, ok := t.transactions[id]
		if !ok {
			return
		}

		tran.Canceled = true
	})
}

func (t *TransactionStateManager) SetResultPath(id string, path string) {
	syncFunc(func() {
		tran, ok := t.transactions[id]
		if !ok {
			return
		}

		tran.ResultPath = path
	})
}

func (t *TransactionStateManager) DeleteTransaction(id string) {
	syncFunc(func() {
		tran, ok := t.transactions[id]
		if !ok {
			return
		}

		os.RemoveAll(tran.Path)
		delete(t.transactions, id)
	})
}

func (t *TransactionStateManager) GetResultPath(id string) (string, error) {
	mu.RLock()
	trans, ok := t.transactions[id]
	mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("transaction doesn't exists")
	}

	if !trans.Done || trans.ResultPath == "" {
		return "", fmt.Errorf("transaction hasn't yet been completed")
	}

	return trans.ResultPath, nil
}

func syncFunc(f func()) {
	mu.Lock()
	f()
	mu.Unlock()
}
