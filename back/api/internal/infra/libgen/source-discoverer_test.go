package libgen

import (
	"context"
	"fmt"
	"ismelen/inkomi/internal/domain/book"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// libgenHTML is a minimal response that passes both Plus and Classic probeCheck.
const libgenHTML = `<html><title>Library Genesis</title><body>
<table id="tablelibgen"><tr><td>Clean Code</td></tr></table>
<table class="c"><tr><td>1</td></tr></table>
</body></html>`

// probeServer creates a test server that responds to /index.php and /search.php
// with real-looking LibGen HTML, and returns 404 for any other path.
func probeServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/index.php" || r.URL.Path == "/search.php" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, libgenHTML)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

// slowProbeServer is like probeServer but adds a delay, used to test that the
// fastest mirror wins the race.
func slowProbeServer(t *testing.T, delay time.Duration) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/index.php" || r.URL.Path == "/search.php" {
			time.Sleep(delay)
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, libgenHTML)
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

// nginxDefaultServer simulates a zombie mirror: returns 200 with nginx welcome
// on GET / but 404 on all real LibGen paths (/index.php, /search.php, /ads.php).
// The old GET / health-check would have selected it; the new probe-based check rejects it.
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
	fast := probeServer(t)
	defer fast.Close()

	slow := slowProbeServer(t, 300*time.Millisecond)
	defer slow.Close()

	broken := brokenServer(t)
	defer broken.Close()

	originalGetMirrors := getMirrors
	getMirrors = func() []book.BooksSource {
		return []book.BooksSource{
			NewPlusMirror(broken.URL),
			NewPlusMirror(slow.URL),
			NewPlusMirror(fast.URL),
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
	// A zombie mirror that answers GET / with 200 (nginx default) but returns
	// 404 on /index.php and /search.php. The old GET / check selected it; the
	// new probe-based check must reject it.
	zombie := nginxDefaultServer(t)
	defer zombie.Close()

	originalGetMirrors := getMirrors
	getMirrors = func() []book.BooksSource {
		return []book.BooksSource{
			NewPlusMirror(zombie.URL),
			NewClassicMirror(zombie.URL),
		}
	}
	defer func() { getMirrors = originalGetMirrors }()

	discoverer := NewSourceDiscoverer()

	source := discoverer.UpdateSource()
	if source != nil {
		t.Errorf("zombie mirror should be rejected, got source: %v", source.GetURL())
	}
}

func TestSourceDiscoverer_UpdateSource_NoValidMirrors_ShouldReturnNil(t *testing.T) {
	broken := brokenServer(t)
	defer broken.Close()

	originalGetMirrors := getMirrors
	getMirrors = func() []book.BooksSource {
		return []book.BooksSource{
			NewPlusMirror(broken.URL),
			NewClassicMirror(broken.URL),
		}
	}
	defer func() { getMirrors = originalGetMirrors }()

	discoverer := NewSourceDiscoverer()

	source := discoverer.UpdateSource()
	if source != nil {
		t.Errorf("expected nil when all mirrors fail, got %v", source)
	}
}

func TestSourceDiscoverer_UpdateSource_ClassicMirrorSelected(t *testing.T) {
	// Verify ClassicMirror (using /search.php) is also picked up correctly.
	srv := probeServer(t)
	defer srv.Close()

	originalGetMirrors := getMirrors
	getMirrors = func() []book.BooksSource {
		return []book.BooksSource{
			NewClassicMirror(srv.URL),
		}
	}
	defer func() { getMirrors = originalGetMirrors }()

	discoverer := NewSourceDiscoverer()

	source := discoverer.UpdateSource()
	if source == nil {
		t.Fatal("expected ClassicMirror to be selected, got nil")
	}
	if source.GetURL() != srv.URL {
		t.Errorf("expected %s, got %s", srv.URL, source.GetURL())
	}
}

func TestSourceDiscoverer_Start_Normal_ShouldCallOnUpdate(t *testing.T) {
	fast := probeServer(t)
	defer fast.Close()

	originalGetMirrors := getMirrors
	getMirrors = func() []book.BooksSource {
		return []book.BooksSource{
			NewPlusMirror(fast.URL),
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
