package state

import (
	"fmt"
	"ismelen/inkomi/internal/domain"
	"os"
	"sync"
	"time"
)

type TransactionManager struct {
	transactions sync.Map
}

var transactionManager *TransactionManager
var once sync.Once

func GetManager() *TransactionManager {
	if transactionManager == nil {
		once.Do(func() {
			transactionManager = &TransactionManager{}
		})
	}
	return transactionManager
}

func (t *TransactionManager) StartTransaction(id, path string, pages int) *domain.Transaction {
	tran := domain.NewTransaction(id, path, pages)
	t.transactions.Store(id, tran)

	time.AfterFunc(90*time.Minute, func() {
		t.DeleteTransaction(id)
	})

	return tran
}

func (t *TransactionManager) getTran(id string) (*domain.Transaction, error) {
	val, ok := t.transactions.Load(id)
	if !ok {
		return nil, fmt.Errorf("transaction doesn't exists")
	}

	tran, ok := val.(*domain.Transaction)
	if !ok {
		return nil, fmt.Errorf("invalid transaction type in map")
	}

	return tran, nil
}

func (t *TransactionManager) CheckProgress(id string) (int, error) {
	tran, err := t.getTran(id)
	if err != nil {
		return 0, err
	}

	if tran.Error != nil {
		return 0, tran.Error
	}

	return tran.ProcessedPages * 100 / tran.Pages, nil
}

func (t *TransactionManager) Cancel(id string) {
	tran, err := t.getTran(id)
	if err != nil {
		return
	}

	tran.Cancel()
}

func (t *TransactionManager) DeleteTransaction(id string) {
	tran, err := t.getTran(id)
	if err != nil {
		return
	}

	os.RemoveAll(tran.Path)
	t.transactions.Delete(id)
}

func (t *TransactionManager) GetResultPath(id string) (string, error) {
	tran, err := t.getTran(id)
	if err != nil {
		return "", err
	}

	return tran.GetResultPath()
}
