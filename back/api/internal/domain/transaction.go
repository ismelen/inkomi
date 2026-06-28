package domain

import (
	"fmt"
	"time"
)

type Transaction struct {
	Id             string
	StartAt        time.Time
	Path           string
	ResultPath     string
	Pages          int
	ProcessedPages int
	Error          error
	Done           bool
	Canceled       bool
}

func NewTransaction(id, path string, pages int) *Transaction {
	return &Transaction{
		Id:      id,
		StartAt: time.Now(),
		Pages:   pages,
		Path:    path,
	}
}

func (t *Transaction) AddProcessedPages(pages int) bool {
	t.ProcessedPages += pages
	return t.Canceled
}

func (t *Transaction) GetProgress() int {
	return t.ProcessedPages * 100 / t.Pages
}

func (t *Transaction) SetDone() {
	t.Done = true
	t.ProcessedPages = t.Pages
}

func (t *Transaction) Cancel() {
	t.Canceled = true
}

func (t *Transaction) SetResultPath(path string) {
	t.ResultPath = path
}

func (t *Transaction) GetResultPath() (string, error) {
	if !t.Done || t.ResultPath == "" {
		return "", fmt.Errorf("transaction hasn't yet been completed")
	}

	return t.ResultPath, nil
}

func (t *Transaction) SetError(err error) {
	t.Error = err
}
