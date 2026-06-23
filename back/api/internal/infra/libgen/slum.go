package libgen

import (
	"encoding/json"
	"ismelen/inkomi/internal/ports"
	"net/http"
	"strings"
	"time"
)

type slumResponse struct {
	PublicGroupList []struct {
		Name        string `json:"name"`
		MonitorList []struct {
			Name    string `json:"name"`
			URL     string `json:"url"`
			SendURL int    `json:"sendUrl"`
		} `json:"monitorList"`
	} `json:"publicGroupList"`
}

var asdfas ports.LibgenMirror = NewClassicMirror("adsf")

var fallbackMirrors = []ports.LibgenMirror{
	NewPlusMirror("https://libgen.bz"),
	NewPlusMirror("https://libgen.la"),
	NewPlusMirror("https://libgen.gl"),
	NewPlusMirror("https://libgen.vg"),
	NewClassicMirror("https://libgen.is"),
	NewClassicMirror("https://libgen.st"),
	NewClassicMirror("https://libgen.rs"),
}

func getMirrors() []ports.LibgenMirror {
	client := &http.Client{Timeout: 8 * time.Second}
	req, _ := http.NewRequest("GET", "https://open-slum.org/api/status-page/slum", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return fallbackMirrors
	}

	var slum slumResponse
	if err := json.NewDecoder(resp.Body).Decode(&slum); err != nil {
		return fallbackMirrors
	}

	var mirrors []ports.LibgenMirror
	for _, group := range slum.PublicGroupList {
		if !strings.Contains(strings.ToLower(group.Name), "libgen") &&
			!strings.Contains(strings.ToLower(group.Name), "library genesis") {
			continue
		}

		for _, m := range group.MonitorList {
			if m.URL == "" || m.SendURL == 0 {
				continue
			}

			base := strings.TrimRight(m.URL, "/")

			var mirror ports.LibgenMirror
			if strings.Contains(m.Name, "+") {
				mirror = NewPlusMirror(base)
			} else {
				mirror = NewClassicMirror(base)
			}

			mirrors = append(mirrors, mirror)
		}
	}

	if len(mirrors) == 0 {
		return fallbackMirrors
	}

	return mirrors
}
