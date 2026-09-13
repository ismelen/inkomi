package libgen

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// slumUpHTML returns a minimal SLUM-style HTML page with the given mirror cards.
// Each entry is (url, status) where status is "up", "protected", "degraded", or "down".
func slumUpHTML(entries []struct{ url, status string }) string {
	html := `<!DOCTYPE html><html><body><div class="container">`
	for i, e := range entries {
		id := fmt.Sprintf("mirror%d", i)
		html += fmt.Sprintf(`
<div class="card" id="%s">
  <div class="card-header">
    <div class="domain-info">
      <a href="%s" class="card-title">%s</a>
    </div>
    <div>
      <span class="status-badge %s">%s</span>
    </div>
  </div>
</div>`, id, e.url, e.url, e.status, e.status)
	}
	html += `</div></body></html>`
	return html
}

func TestSlum_GetMirrors_ServerError_ShouldReturnNil(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	original := slumURL
	slumURL = server.URL
	defer func() { slumURL = original }()

	mirrors := getMirrors()
	if mirrors != nil {
		t.Errorf("expected nil when server returns 500, got %d mirrors", len(mirrors))
	}
}

func TestSlum_GetMirrors_OnlyUpMirrors_ShouldReturnThem(t *testing.T) {
	html := slumUpHTML([]struct{ url, status string }{
		{"https://libgen.bz", "up"},
		{"https://libgen.vg", "up"},
		{"https://libgen.gl", "protected"},
		{"https://libgen.la", "protected"},
		{"https://libgen.li", "degraded"},
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, html)
	}))
	defer server.Close()

	original := slumURL
	slumURL = server.URL
	defer func() { slumURL = original }()

	mirrors := getMirrors()

	// Each UP domain produces a PlusMirror and a ClassicMirror candidate.
	expectedCount := 2 * 2 // 2 UP domains × 2 mirror types
	if len(mirrors) != expectedCount {
		t.Fatalf("expected %d mirrors (2 per UP domain), got %d", expectedCount, len(mirrors))
	}

	// Verify both UP domains are represented.
	urls := make(map[string]int)
	for _, m := range mirrors {
		urls[m.GetURL()]++
	}
	for _, expected := range []string{"https://libgen.bz", "https://libgen.vg"} {
		if urls[expected] != 2 {
			t.Errorf("expected 2 entries for %s (Plus+Classic), got %d", expected, urls[expected])
		}
	}
	for _, excluded := range []string{"https://libgen.gl", "https://libgen.la", "https://libgen.li"} {
		if urls[excluded] > 0 {
			t.Errorf("excluded mirror %s should not be present", excluded)
		}
	}
}

func TestSlum_GetMirrors_NoUpMirrors_ShouldReturnNil(t *testing.T) {
	html := slumUpHTML([]struct{ url, status string }{
		{"https://libgen.gl", "protected"},
		{"https://libgen.la", "protected"},
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, html)
	}))
	defer server.Close()

	original := slumURL
	slumURL = server.URL
	defer func() { slumURL = original }()

	mirrors := getMirrors()
	if len(mirrors) != 0 {
		t.Errorf("expected 0 mirrors when all are protected, got %d", len(mirrors))
	}
}

func TestSlum_GetMirrors_TrimsTrailingSlash(t *testing.T) {
	html := slumUpHTML([]struct{ url, status string }{
		{"https://libgen.bz/", "up"}, // trailing slash
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, html)
	}))
	defer server.Close()

	original := slumURL
	slumURL = server.URL
	defer func() { slumURL = original }()

	mirrors := getMirrors()
	for _, m := range mirrors {
		if url := m.GetURL(); url[len(url)-1] == '/' {
			t.Errorf("mirror URL should not have trailing slash: %s", url)
		}
	}
}
