package libgen

import (
	"context"
	"sync"
	"time"
)

const pingTimeout = 12 * time.Second

// fastestMirror pings every mirror at once and returns the first one that answers 200.
func fastestMirror(ctx context.Context, mirrors []IMirror) (IMirror, bool) {
	ctx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	winner := make(chan IMirror, 1)
	var wg sync.WaitGroup
	for _, m := range mirrors {
		wg.Go(func() {
			if ping(ctx, m.URL()) {
				select {
				case winner <- m:
				default:
				}
			}
		})
	}

	allDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(allDone)
	}()

	select {
	case m := <-winner:
		return m, true
	case <-allDone:
		select {
		case m := <-winner:
			return m, true
		default:
			return nil, false
		}
	case <-ctx.Done():
		return nil, false
	}
}

func ping(ctx context.Context, baseURL string) bool {
	resp, err := get(ctx, pingClient, baseURL+"/")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}
