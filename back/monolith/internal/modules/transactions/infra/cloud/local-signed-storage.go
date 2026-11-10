package cloud

import (
	"net/http"
	"time"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/ports"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/storage"
)

var _ ports.ICloudStorage = (*LocalSignedStorage)(nil)

// LocalSignedStorage hands out urls of the monolith's own /files endpoint, signed like
// presigned R2 urls, for the files kept on the server's disk.
type LocalSignedStorage struct {
	store  *storage.LocalStorage
	signer *storage.URLSigner
}

func NewLocalSignedStorage(store *storage.LocalStorage, signer *storage.URLSigner) *LocalSignedStorage {
	return &LocalSignedStorage{store: store, signer: signer}
}

// GetUrl signs a PUT that accepts at most size bytes, the space reserved for the file; without a
// declared size the body isn't limited.
func (l *LocalSignedStorage) GetUrl(id string, filename string, size int64) (string, error) {
	limit := size
	if limit <= 0 {
		limit = -1
	}
	url, _ := l.signer.URL(http.MethodPut, id, filename, presignExpiry, limit)
	return url, nil
}

func (l *LocalSignedStorage) GetDownloadUrl(id string, filename string) (string, time.Time, error) {
	url, expiresAt := l.signer.URL(http.MethodGet, id, filename, presignExpiry, -1)
	return url, expiresAt, nil
}

func (l *LocalSignedStorage) Check(id string, filename string) bool {
	return l.store.Exists(id, filename)
}
