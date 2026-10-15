//go:build e2e

package e2e_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const flowTimeout = 8 * time.Minute

type sourceEvent struct {
	Id       string `json:"id"`
	Status   string `json:"status"`
	Filename string `json:"filename"`
	Error    string `json:"error"`
}

type uploadSource struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Filename string `json:"filename"`
	Url      string `json:"url"`
}

// TestMangaUpload_EndToEnd goes through every service like the app does: it registers a user
// (auth-service), asks transaction-manager for upload urls, uploads two chapters to R2 and
// follows their events over the websocket until push-service marks them as sent. One chapter
// becomes an epub (manga-worker) and the other a kepub (manga-worker → epub-worker).
func TestMangaUpload_EndToEnd(t *testing.T) {
	t.Parallel()

	// Arrange
	e := loadEnv(t)
	storage := newCloudStorage(t)
	token := registerUser(t, e)
	events := subscribe(t, e, token)
	cbz := buildCBZ(t)
	stamp := time.Now().Unix()

	request := map[string]any{
		"ereader_key": "K1",
		"config": map[string]any{
			"epub":  map[string]any{"ereaderKey": "K1", "grayscale": true},
			"kepub": map[string]any{"ereaderKey": "KoMT", "grayscale": true, "dithering": true, "palette8bit": true},
		},
		"sources": []map[string]any{
			{"title": fmt.Sprintf("E2E epub %d", stamp), "type": "file", "filename": "e2e-epub.cbz", "size": len(cbz), "config_hash": "epub", "reading_direction": "rtl"},
			{"title": fmt.Sprintf("E2E kepub %d", stamp), "type": "file", "filename": "e2e-kepub.cbz", "size": len(cbz), "config_hash": "kepub", "reading_direction": "ltr"},
		},
	}
	wantExt := map[string]string{
		fmt.Sprintf("E2E epub %d", stamp):  ".epub",
		fmt.Sprintf("E2E kepub %d", stamp): ".kepub.epub",
	}

	var sources []uploadSource
	status := doJSON(t, http.MethodPost, e.tmURL+"/uploads/request", token, request, &sources)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, sources, 2)

	results := map[string]string{}
	t.Cleanup(func() {
		ctx := context.Background()
		for _, src := range sources {
			// best effort: the sandbox shares the production bucket, so leave nothing behind
			_ = storage.Delete(ctx, src.Id, src.Filename)
			if filename, ok := results[src.Id]; ok {
				_ = storage.Delete(ctx, src.Id, filename)
			}
		}
	})

	// Act
	for _, src := range sources {
		require.NotEmpty(t, src.Url, "upload url of %s", src.Title)
		putFile(t, src.Url, cbz)
		status := doJSON(t, http.MethodPatch, e.tmURL+"/uploads/done/"+src.Id, token, nil, nil)
		require.Less(t, status, 300, "upload done of %s", src.Title)
	}
	results = awaitSent(t, events, sources)

	// Assert
	for _, src := range sources {
		filename := results[src.Id]
		assert.True(t, strings.HasSuffix(filename, wantExt[src.Title]), "%s: result %q", src.Title, filename)
		if wantExt[src.Title] == ".epub" {
			assert.False(t, strings.HasSuffix(filename, ".kepub.epub"), "%s: unexpected kepub", src.Title)
		}

		size, err := storage.Size(context.Background(), src.Id, filename)
		require.NoError(t, err, "%s: result missing in R2", src.Title)
		assert.Positive(t, size)

		_, err = storage.Size(context.Background(), src.Id, src.Filename)
		assert.Error(t, err, "%s: the source cbz should have been deleted", src.Title)
	}
}

// subscribe opens the transaction-manager websocket and forwards the source events.
func subscribe(t *testing.T, e env, token string) <-chan sourceEvent {
	t.Helper()

	wsURL := "ws" + strings.TrimPrefix(e.tmURL, "http") + "/ws"
	header := http.Header{"Authorization": {"Bearer " + token}}
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if resp != nil {
		resp.Body.Close()
	}
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	events := make(chan sourceEvent, 64)
	go func() {
		defer close(events)
		for {
			var event sourceEvent
			if err := conn.ReadJSON(&event); err != nil {
				return
			}
			events <- event
		}
	}()
	return events
}

// awaitSent waits until every source is sent and returns their result filenames by id. A failed
// source fails the test right away with its error.
func awaitSent(t *testing.T, events <-chan sourceEvent, sources []uploadSource) map[string]string {
	t.Helper()

	pending := map[string]string{}
	for _, src := range sources {
		pending[src.Id] = src.Title
	}
	filenames := map[string]string{}
	timeout := time.After(flowTimeout)

	for len(pending) > 0 {
		select {
		case event, ok := <-events:
			require.True(t, ok, "websocket closed while waiting for %v", pending)
			title, tracked := pending[event.Id]
			if !tracked {
				continue
			}
			t.Logf("%s: %s %s", title, event.Status, event.Filename)
			if event.Filename != "" {
				filenames[event.Id] = event.Filename
			}
			switch event.Status {
			case "failed":
				require.FailNow(t, "source failed", "%s: %s", title, event.Error)
			case "sent":
				delete(pending, event.Id)
			}
		case <-timeout:
			require.FailNow(t, "timeout", "still waiting for %v", pending)
		}
	}
	return filenames
}
