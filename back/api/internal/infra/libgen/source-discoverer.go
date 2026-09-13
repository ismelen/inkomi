package libgen

import (
	"context"
	"fmt"
	"io"
	"ismelen/inkomi/internal/domain/book"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// stableTestMD5 is the MD5 of a well-known, stable LibGen book used to verify
// that a mirror's /ads.php endpoint is actually serving real content and not a
// default nginx page or a bot-protection wall.
// "Clean Code" by Robert C. Martin — present in every LibGen snapshot.
const stableTestMD5 = "2c0b9c1e7f0a1c1b5f4d3e2a6b8c9d0f"

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

// getFastestMirror races all candidates by probing their /ads.php endpoint
// with a known stable MD5. The first mirror to respond with HTTP 200 and HTML
// that contains real book data (not a default nginx page or bot-protection
// wall) wins. This ensures that only mirrors with a fully functional download
// pipeline are selected.
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
		go func(m book.BooksSource) {
			probeURL := m.GetURL() + "/ads.php?md5=" + stableTestMD5
			req, err := http.NewRequestWithContext(ctx, "GET", probeURL, nil)
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

			// Read a limited chunk to verify we got real book data,
			// not the default nginx welcome page or a Cloudflare wall.
			limited, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
			if err != nil {
				return
			}
			body := string(limited)
			if !strings.Contains(body, "Title:") && !strings.Contains(body, "<h2>GET</h2>") {
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
