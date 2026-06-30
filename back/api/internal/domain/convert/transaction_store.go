package convert

// TransactionStore is the port implemented by infra/store to manage in-memory transactions.
type TransactionStore interface {
	StartTransaction(id, path string, pages int) *Transaction
	DeleteTransaction(id string)
	CheckProgress(id string) (int, error)
	Cancel(id string)
	GetResultPath(id string) (string, error)
}
