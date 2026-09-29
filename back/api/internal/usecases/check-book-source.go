package usecases

import (
	"context"
	"ismelen/inkomi/internal/domain/book"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
)

type CheckBooksSourceUC struct {
	provider  book.BooksProvider
	mu        sync.RWMutex
	lastCheck time.Time
}

func NewCheckBooksSourceUC(provider book.BooksProvider) *CheckBooksSourceUC {
	return &CheckBooksSourceUC{
		provider: provider,
	}
}

func (c *CheckBooksSourceUC) Execute(ctx context.Context) (bool, error) {
	ctx, span := otel.Tracer("inkomi-api").Start(ctx, "CheckBooksSourceUC.Execute")
	defer span.End()

	mirror, ok := c.provider.GetMirror()
	if ok && c.pingMirror(ctx, mirror) {
		return true, nil
	}

	c.mu.RLock()
	isCooldown := time.Since(c.lastCheck) < time.Hour
	c.mu.RUnlock()

	if isCooldown {
		return false, nil
	}

	c.mu.Lock()
	c.lastCheck = time.Now()
	c.mu.Unlock()

	if updated := c.provider.Refresh(); !updated {
		return false, nil
	}

	mirror, ok = c.provider.GetMirror()
	if ok {
		return true, nil
	}

	return false, nil
}

func (c *CheckBooksSourceUC) pingMirror(ctx context.Context, m book.BooksSource) bool {
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", m.GetURL()+"/", nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
