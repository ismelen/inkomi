package libgen_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ismelen/inkomi/back/library-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/library-service/internal/infra/libgen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testMD5 models.MD5 = "0123456789abcdef0123456789abcdef"

const plusSearchPage = `<html><body><table id="tablelibgen">
<tr><th>cover</th><th>title</th><th>author</th><th>pub</th><th>year</th><th>lang</th><th>pages</th><th>size</th><th>ext</th><th>mirrors</th></tr>
<tr>
  <td><img src="/covers/1.jpg"></td>
  <td><a href="series.php?id=1">Series</a><a href="edition.php?id=9">Dune: Deluxe Edition</a></td>
  <td>Frank Herbert</td><td>Ace</td><td>1965</td><td>English</td><td>412</td><td>1 MB</td><td>EPUB</td>
  <td><a href="/ads.php?md5=0123456789ABCDEF0123456789ABCDEF">libgen</a></td>
</tr>
<tr>
  <td></td><td>No pages</td><td>x</td><td></td><td></td><td>English</td><td>?</td><td></td><td>pdf</td><td></td>
</tr>
</table></body></html>`

func adsPage(href string) string {
	return `<html><body>Title: Dune &amp; Messiah<br>Author: Frank Herbert<br>
Extension: epub, Size: 1 MB<br><a href="` + href + `"><h2>GET</h2></a></body></html>`
}

func newPlusServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/index.php", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "dune", r.URL.Query().Get("req"))
		fmt.Fprint(w, plusSearchPage)
	})
	mux.HandleFunc("/ads.php", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("md5") {
		case string(testMD5):
			fmt.Fprint(w, adsPage("get.php?md5="+string(testMD5)+"&amp;key=K"))
		default:
			fmt.Fprint(w, "<html><body>nothing here</body></html>")
		}
	})
	mux.HandleFunc("/get.php", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "K", r.URL.Query().Get("key"))
		fmt.Fprint(w, "epub bytes")
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func TestPlusMirror_Search_ResultsPage_ParsesBooks(t *testing.T) {
	t.Parallel()

	// Arrange
	server := newPlusServer(t)
	mirror := libgen.NewPlusMirror(server.URL)

	// Act
	books, err := mirror.Search(context.Background(), "dune")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []models.Book{{
		Title:     "Dune: Deluxe Edition",
		Author:    "Frank Herbert",
		Pages:     412,
		Language:  "English",
		Extension: "epub",
		MD5:       string(testMD5),
		CoverURL:  server.URL + "/covers/1.jpg",
	}}, books)
}

func TestMirror_Download_KnownMD5_StreamsBook(t *testing.T) {
	t.Parallel()

	// Arrange
	server := newPlusServer(t)
	mirror := libgen.NewPlusMirror(server.URL)

	// Act
	book, err := mirror.Download(context.Background(), testMD5)

	// Assert
	require.NoError(t, err)
	defer book.Stream.Close()
	content, err := io.ReadAll(book.Stream)
	require.NoError(t, err)
	assert.Equal(t, "epub bytes", string(content))
	assert.Equal(t, "Dune & Messiah", book.Title)
	assert.Equal(t, "epub", book.Ext)
}

func TestMirror_Download_MissingBook_ReturnsErrBookNotFound(t *testing.T) {
	t.Parallel()

	// Arrange
	server := newPlusServer(t)
	mirror := libgen.NewClassicMirror(server.URL)

	// Act
	book, err := mirror.Download(context.Background(), "ffffffffffffffffffffffffffffffff")

	// Assert
	assert.ErrorIs(t, err, models.ErrBookNotFound)
	assert.Nil(t, book)
}

func TestMirror_Download_ServerError_ReturnsTransientError(t *testing.T) {
	t.Parallel()

	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)
	mirror := libgen.NewPlusMirror(server.URL)

	// Act
	_, err := mirror.Download(context.Background(), testMD5)

	// Assert
	require.Error(t, err)
	assert.NotErrorIs(t, err, models.ErrBookNotFound)
}

func TestClassicMirror_Search_ResultsPage_ReadsMetadata(t *testing.T) {
	t.Parallel()

	// Arrange
	mux := http.NewServeMux()
	mux.HandleFunc("/search.php", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<table class="c"><tr><td>ID</td></tr><tr><td>11</td></tr><tr><td>12</td></tr><tr><td>11</td></tr></table>`)
	})
	mux.HandleFunc("/json.php", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "11,12", r.URL.Query().Get("ids"))
		fmt.Fprint(w, `[{"title":"Dune ","author":"Frank Herbert","pages":"412","language":"English","extension":"EPUB","md5":"0123456789ABCDEF0123456789ABCDEF"},
{"title":"Dune Messiah","author":"Frank Herbert","pages":"","language":"English","extension":"pdf","md5":""}]`)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	// Act
	books, err := libgen.NewClassicMirror(server.URL).Search(context.Background(), "dune")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []models.Book{
		{Title: "Dune", Author: "Frank Herbert", Pages: 412, Language: "English", Extension: "epub", MD5: string(testMD5)},
		{Title: "Dune Messiah", Author: "Frank Herbert", Language: "English", Extension: "pdf"},
	}, books)
}

func TestClassicMirror_Search_NoIds_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<table class="c"><tr><td>ID</td></tr></table>`)
	}))
	t.Cleanup(server.Close)

	// Act
	books, err := libgen.NewClassicMirror(server.URL).Search(context.Background(), "nothing")

	// Assert
	require.NoError(t, err)
	assert.Empty(t, books)
}

func TestSlumMirrorList_Mirrors_StatusPage_ReturnsLibgenMirrors(t *testing.T) {
	t.Parallel()

	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"publicGroupList": []map[string]any{
				{"name": "Library Genesis", "monitorList": []map[string]any{
					{"name": "Classic", "url": "http://libgen.is/", "sendUrl": 1},
					{"name": "Plus +", "url": "http://libgen.la/", "sendUrl": 1},
					{"name": "Empty", "url": "", "sendUrl": 1},
					{"name": "Hidden", "url": "http://libgen.ns/", "sendUrl": 0},
				}},
				{"name": "Other", "monitorList": []map[string]any{
					{"name": "Other", "url": "http://other.com/", "sendUrl": 1},
				}},
			},
		}))
	}))
	t.Cleanup(server.Close)
	list := libgen.NewSlumMirrorList(server.URL, libgen.FallbackMirrors())

	// Act
	mirrors := list.Mirrors(context.Background())

	// Assert
	require.Len(t, mirrors, 2)
	assert.IsType(t, &libgen.ClassicMirror{}, mirrors[0])
	assert.Equal(t, "http://libgen.is", mirrors[0].URL())
	assert.IsType(t, &libgen.PlusMirror{}, mirrors[1])
	assert.Equal(t, "http://libgen.la", mirrors[1].URL())
}

func TestSlumMirrorList_Mirrors_StatusPageFails_ReturnsFallback(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{"Server error", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }},
		{"Invalid JSON", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "{") }},
		{"No libgen group", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"publicGroupList":[]}`) }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			server := httptest.NewServer(tc.handler)
			t.Cleanup(server.Close)
			fallback := []libgen.IMirror{libgen.NewPlusMirror("http://fallback")}

			// Act
			mirrors := libgen.NewSlumMirrorList(server.URL, fallback).Mirrors(context.Background())

			// Assert
			assert.Equal(t, fallback, mirrors)
		})
	}
}
