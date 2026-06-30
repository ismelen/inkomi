package libgen

import (
	"context"
	"ismelen/inkomi/internal/domain/book"
	"log"
	"net/http"
	"sync"
	"time"
)

func (l *LibgenService) StartDiscovery(ctx context.Context, interval time.Duration) {
	l.updateMirror()
	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ticker.C:
				l.updateMirror()
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (l *LibgenService) updateMirror() {
	mirrors := getMirrors()

	fastest, ok := l.getFastestMirror(mirrors)
	if !ok {
		log.Println("Could't update mirror")
		return
	}

	log.Printf("\nNew mirror: %s", fastest.GetURL())
	l.mirror.Store(fastest)
}

func (l *LibgenService) getFastestMirror(mirrors []book.LibgenMirror) (book.LibgenMirror, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	winner := make(chan book.LibgenMirror, 1)
	var once sync.Once

	client := &http.Client{
		Timeout: 12 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for _, m := range mirrors {
		go func(m book.LibgenMirror) {
			req, err := http.NewRequestWithContext(ctx, "GET", m.GetURL()+"/", nil)
			if err != nil {
				return
			}

			req.Header.Set("User-Agent", "Mozilla/5.0")
			resp, err := client.Do(req)
			if err != nil {
				return
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return
			}

			once.Do(func() { winner <- m })
		}(m)
	}

	select {
	case m := <-winner:
		return m, true
	case <-ctx.Done():
		return nil, false
	}
}
