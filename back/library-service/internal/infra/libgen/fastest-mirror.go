package libgen

import (
	"context"
	"time"
)

const pingTimeout = 12 * time.Second

// fastestMirror pings every mirror at once and returns the first one that answers 200.
func fastestMirror(ctx context.Context, mirrors []IMirror) (IMirror, bool) {
	ctx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	results := make(chan IMirror, len(mirrors))
	for _, m := range mirrors {
		go func() {
			if ping(ctx, m.URL()) {
				results <- m
				return
			}
			results <- nil
		}()
	}

	for range mirrors {
		if m := <-results; m != nil {
			return m, true
		}
	}
	return nil, false
}

func ping(ctx context.Context, baseURL string) bool {
	resp, err := get(ctx, pingClient, baseURL+"/")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}
