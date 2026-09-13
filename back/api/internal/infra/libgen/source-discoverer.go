package libgen

import (
	"context"
	"fmt"
	"io"
	"ismelen/inkomi/internal/domain/book"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// mirrorProber is implemented by mirror types that expose a type-appropriate
// probe endpoint for health-checking (PlusMirror → /index.php, ClassicMirror →
// /search.php). This avoids relying on a specific book MD5 being present.
type mirrorProber interface {
	probeURL() string
	probeCheck(body string) bool
}

// getFastestMirror races all candidates by probing their type-appropriate
// search endpoint. The first mirror that responds with HTTP 200 and real LibGen
// HTML (not nginx default, not a bot-protection wall) wins.
func (s *SourceDiscoverer) getFastestMirror(mirrors []book.BooksSource) (book.BooksSource, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	winner := make(chan book.BooksSource, 1)
	var once sync.Once

	client := &http.Client{
		Timeout:   15 * time.Second,
		Transport: baseTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for _, m := range mirrors {
		prober, ok := m.(mirrorProber)
		if !ok {
			continue
		}

		go func(m book.BooksSource, prober mirrorProber) {
			req, err := http.NewRequestWithContext(ctx, "GET", prober.probeURL(), nil)
			if err != nil {
				return
			}
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
			req.Header.Set("Accept", "text/html,application/xhtml+xml,*/*;q=0.8")
			req.Header.Set("Accept-Encoding", "identity")

			resp, err := client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return
			}

			limited, err := io.ReadAll(io.LimitReader(resp.Body, 8192))
			if err != nil {
				return
			}

			if !prober.probeCheck(string(limited)) {
				return
			}

			once.Do(func() { winner <- m })
		}(m, prober)
	}

	select {
	case m := <-winner:
		return m, true
	case <-ctx.Done():
		return nil, false
	}
}

type SourceDiscoverer struct {
	singleUpdater singleflight.Group
	lastCheckMu   sync.RWMutex
	lastCheck     time.Time
	onUpdate      func(book.BooksSource)
}

func NewSourceDiscoverer() *SourceDiscoverer {
	return &SourceDiscoverer{}
}

func (s *SourceDiscoverer) Start(ctx context.Context, interval time.Duration) {
	go func() {
		if source := s.UpdateSource(); source != nil {
			s.onUpdate(source)
		}

		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ticker.C:
				if source := s.UpdateSource(); source != nil {
					s.onUpdate(source)
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *SourceDiscoverer) SetOnUpdate(onUpdate func(book.BooksSource)) {
	s.onUpdate = onUpdate
}

func (s *SourceDiscoverer) UpdateSource() book.BooksSource {
	mirror, err, _ := s.singleUpdater.Do("refresh", func() (any, error) {
		mirrors := getMirrors()

		fastest, ok := s.getFastestMirror(mirrors)

		s.lastCheckMu.Lock()
		s.lastCheck = time.Now()
		s.lastCheckMu.Unlock()

		if !ok {
			return nil, fmt.Errorf("Couldn't update mirror")
		}

		log.Printf("New mirror: %s", fastest.GetURL())
		return fastest, nil
	})

	if err != nil {
		log.Print(err.Error())
		return nil
	}

	return mirror.(book.BooksSource)
}
