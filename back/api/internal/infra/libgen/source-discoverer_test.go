package libgen

import (
	"context"
	"fmt"
	"ismelen/inkomi/internal/domain/book"
	"ismelen/inkomi/internal/test/mocks"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// adsHTML returns a minimal ads.php-style response that passes the health-check.
const adsHTML = `<html><body>
<td>Title: <b>Clean Code</b></td>
<td>Extension: <b>epub</b></td>
<a href="/get/cleancode.epub"><h2>GET</h2></a>
</body></html>`

// adsServer creates a test server that responds to /ads.php with real-looking
// book data and any other path with 404.
func adsServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ads.php" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, adsHTML)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

// brokenServer returns 500 on all paths (simulates a genuinely down mirror).
func brokenServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
}

// nginxDefaultServer simulates a mirror with unconfigured nginx: returns 200 on /
// but 404 on /ads.php — the classic "zombie mirror" that used to fool the old
// GET / health-check.
func nginxDefaultServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, "<html><body><h1>Welcome to nginx!</h1></body></html>")
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestSourceDiscoverer_UpdateSource_Normal_ShouldReturnFastest(t *testing.T) {
	fast := adsServer(t)
	defer fast.Close()

	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ads.php" {
			time.Sleep(200 * time.Millisecond)
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, adsHTML)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer slow.Close()

	broken := brokenServer(t)
	defer broken.Close()

	originalGetMirrors := getMirrors
	getMirrors = func() []book.BooksSource {
		return []book.BooksSource{
			&mocks.BooksSourceMock{URL: broken.URL},
			&mocks.BooksSourceMock{URL: slow.URL},
			&mocks.BooksSourceMock{URL: fast.URL},
		}
	}
	defer func() { getMirrors = originalGetMirrors }()

	discoverer := NewSourceDiscoverer()

	source := discoverer.UpdateSource()
	if source == nil {
		t.Fatal("expected to find a source, got nil")
	}
	if source.GetURL() != fast.URL {
		t.Errorf("expected fastest server URL %s, got %s", fast.URL, source.GetURL())
	}
}

func TestSourceDiscoverer_UpdateSource_ZombieMirror_ShouldBeRejected(t *testing.T) {
	// A mirror that answers GET / with 200 (nginx default) but /ads.php with 404.
	// With the old health-check it would have been selected; now it must be rejected.
	zombie := nginxDefaultServer(t)
	defer zombie.Close()

	originalGetMirrors := getMirrors
	getMirrors = func() []book.BooksSource {
		return []book.BooksSource{
			&mocks.BooksSourceMock{URL: zombie.URL},
		}
	}
	defer func() { getMirrors = originalGetMirrors }()

	discoverer := NewSourceDiscoverer()

	source := discoverer.UpdateSource()
	if source != nil {
		t.Errorf("zombie mirror (nginx default) should be rejected, but got source: %v", source.GetURL())
	}
}

func TestSourceDiscoverer_UpdateSource_NoValidMirrors_ShouldReturnNil(t *testing.T) {
	broken := brokenServer(t)
	defer broken.Close()

	originalGetMirrors := getMirrors
	getMirrors = func() []book.BooksSource {
		return []book.BooksSource{
			&mocks.BooksSourceMock{URL: broken.URL},
		}
	}
	defer func() { getMirrors = originalGetMirrors }()

	discoverer := NewSourceDiscoverer()

	source := discoverer.UpdateSource()
	if source != nil {
		t.Errorf("expected nil source when all mirrors fail, got %v", source)
	}
}

func TestSourceDiscoverer_Start_Normal_ShouldCallOnUpdate(t *testing.T) {
	fast := adsServer(t)
	defer fast.Close()

	originalGetMirrors := getMirrors
	getMirrors = func() []book.BooksSource {
		return []book.BooksSource{
			&mocks.BooksSourceMock{URL: fast.URL},
		}
	}
	defer func() { getMirrors = originalGetMirrors }()

	discoverer := NewSourceDiscoverer()

	called := make(chan bool, 1)
	discoverer.SetOnUpdate(func(source book.BooksSource) {
		called <- true
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	discoverer.Start(ctx, 100*time.Millisecond)

	select {
	case <-called:
		// Success
	case <-time.After(5 * time.Second):
		t.Error("onUpdate was not called")
	}
}
