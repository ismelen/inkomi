package libgen

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"
)

const (
	SlumURL     = "https://open-slum.org/api/status-page/slum"
	slumTimeout = 8 * time.Second
)

type slumResponse struct {
	PublicGroupList []slumGroup `json:"publicGroupList"`
}

type slumGroup struct {
	Name        string        `json:"name"`
	MonitorList []slumMonitor `json:"monitorList"`
}

type slumMonitor struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	SendURL int    `json:"sendUrl"`
}

func FallbackMirrors() []IMirror {
	return []IMirror{
		NewPlusMirror("https://libgen.bz"),
		NewPlusMirror("https://libgen.la"),
		NewPlusMirror("https://libgen.gl"),
		NewPlusMirror("https://libgen.vg"),
		NewClassicMirror("https://libgen.is"),
		NewClassicMirror("https://libgen.st"),
		NewClassicMirror("https://libgen.rs"),
	}
}

// SlumMirrorList reads the libgen mirrors from the open-slum status page and uses fallback
// when it can't.
type SlumMirrorList struct {
	url      string
	fallback []IMirror
}

func NewSlumMirrorList(url string, fallback []IMirror) *SlumMirrorList {
	return &SlumMirrorList{url: url, fallback: fallback}
}

func (s *SlumMirrorList) Mirrors(ctx context.Context) []IMirror {
	mirrors, err := s.fetch(ctx)
	if err != nil {
		log.Printf("libgen: slum mirrors: %v; using fallback", err)
	}
	if len(mirrors) == 0 {
		return slices.Clone(s.fallback)
	}
	return mirrors
}

func (s *SlumMirrorList) fetch(ctx context.Context) ([]IMirror, error) {
	ctx, cancel := context.WithTimeout(ctx, slumTimeout)
	defer cancel()

	resp, err := get(ctx, pageClient, s.url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var slum slumResponse
	if err := json.NewDecoder(resp.Body).Decode(&slum); err != nil {
		return nil, fmt.Errorf("decode %s: %w", s.url, err)
	}

	var mirrors []IMirror
	for _, group := range slum.PublicGroupList {
		name := strings.ToLower(group.Name)
		if !strings.Contains(name, "libgen") && !strings.Contains(name, "library genesis") {
			continue
		}

		for _, m := range group.MonitorList {
			if m.URL == "" || m.SendURL == 0 {
				continue
			}
			base := strings.TrimRight(m.URL, "/")
			if strings.Contains(m.Name, "+") {
				mirrors = append(mirrors, NewPlusMirror(base))
			} else {
				mirrors = append(mirrors, NewClassicMirror(base))
			}
		}
	}
	return mirrors, nil
}
