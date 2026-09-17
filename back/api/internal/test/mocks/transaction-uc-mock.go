package mocks

import (
	"ismelen/inkomi/internal/domain/convert"
	"ismelen/inkomi/internal/domain/manga"
)

type MockTransactionUC struct {
	ProcessResult *convert.TransactionResultFile
	ExecuteCalled bool
}

func (m *MockTransactionUC) Process(file *convert.TransactionFile, tran *convert.Transaction, transPath string) *convert.TransactionResultFile {
	return m.ProcessResult
}

func (m *MockTransactionUC) Execute(file *convert.TransactionFile, tran *convert.Transaction, transPath string) {
	m.ExecuteCalled = true
}

func (m *MockTransactionUC) GetImageSettings() *manga.ImageSettings {
	return nil
}
