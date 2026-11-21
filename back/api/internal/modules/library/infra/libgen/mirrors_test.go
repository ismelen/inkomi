package libgen_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/infra/libgen"
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
	result, err := mirror.Search(context.Background(), "dune", models.SearchFieldTitle, 1)

	// Assert
	require.NoError(t, err)
	assert.Zero(t, result.NextPage, "no paginator: a single page")
	assert.Equal(t, []models.Book{{
		Title:     "Dune: Deluxe Edition",
		Author:    "Frank Herbert",
		Pages:     412,
		Language:  "English",
		Extension: "epub",
		MD5:       string(testMD5),
		CoverURL:  server.URL + "/covers/1.jpg",
	}}, result.Books)
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
	books, err := libgen.NewClassicMirror(server.URL).Search(context.Background(), "dune", models.SearchFieldTitle, 1)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []models.Book{
		{Title: "Dune", Author: "Frank Herbert", Pages: 412, Language: "English", Extension: "epub", MD5: string(testMD5)},
		{Title: "Dune Messiah", Author: "Frank Herbert", Language: "English", Extension: "pdf"},
	}, books.Books)
	assert.Zero(t, books.NextPage)
}

func TestClassicMirror_Search_NoIds_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<table class="c"><tr><td>ID</td></tr></table>`)
	}))
	t.Cleanup(server.Close)

	// Act
	books, err := libgen.NewClassicMirror(server.URL).Search(context.Background(), "nothing", models.SearchFieldTitle, 1)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, books.Books)
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

func TestPlusMirror_Search_Field_SearchesItsColumn(t *testing.T) {
	t.Parallel()

	cases := []struct {
		field  models.SearchField
		column string
	}{
		{models.SearchFieldTitle, "t"},
		{models.SearchFieldAuthor, "a"},
	}

	for _, tc := range cases {
		t.Run(string(tc.field), func(t *testing.T) {
			t.Parallel()

			// Arrange
			var columns []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				columns = r.URL.Query()["columns[]"]
				assert.Equal(t, "100", r.URL.Query().Get("res"))
				fmt.Fprint(w, plusSearchPage)
			}))
			t.Cleanup(server.Close)

			// Act
			_, err := libgen.NewPlusMirror(server.URL).Search(context.Background(), "mola", tc.field, 1)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, []string{tc.column}, columns)
		})
	}
}

func TestPlusMirror_Search_LazyCover_ReadsDataSrc(t *testing.T) {
	t.Parallel()

	// Arrange
	page := strings.Replace(plusSearchPage, `<img src="/covers/1.jpg">`,
		`<img src="/img/blank.png" data-src="/fictioncovers/1/abc_small.jpg">`, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, page)
	}))
	t.Cleanup(server.Close)

	// Act
	books, err := libgen.NewPlusMirror(server.URL).Search(context.Background(), "dune", models.SearchFieldTitle, 1)

	// Assert
	require.NoError(t, err)
	require.Len(t, books.Books, 1)
	assert.Equal(t, server.URL+"/fictioncovers/1/abc_small.jpg", books.Books[0].CoverURL)
}

func TestClassicMirror_Search_AuthorWithCover_SearchesAuthorAndBuildsCoverURL(t *testing.T) {
	t.Parallel()

	// Arrange
	var column string
	mux := http.NewServeMux()
	mux.HandleFunc("/search.php", func(w http.ResponseWriter, r *http.Request) {
		column = r.URL.Query().Get("column")
		fmt.Fprint(w, `<table class="c"><tr><td>ID</td></tr><tr><td>11</td></tr></table>`)
	})
	mux.HandleFunc("/json.php", func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Query().Get("fields"), "coverurl")
		fmt.Fprint(w, `[{"title":"La nena","author":"Carmen Mola","extension":"epub","md5":"0123456789abcdef0123456789abcdef","coverurl":"1234000/abc.jpg"}]`)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	// Act
	books, err := libgen.NewClassicMirror(server.URL).Search(context.Background(), "mola", models.SearchFieldAuthor, 1)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "author", column)
	require.Len(t, books.Books, 1)
	assert.Equal(t, server.URL+"/covers/1234000/abc.jpg", books.Books[0].CoverURL)
}

func TestMirror_Download_StaleKey_RetriesWithAFreshOne(t *testing.T) {
	t.Parallel()

	// Arrange
	var adsCalls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/ads.php", func(w http.ResponseWriter, r *http.Request) {
		key := fmt.Sprint("K", adsCalls.Add(1))
		fmt.Fprint(w, adsPage("get.php?md5="+string(testMD5)+"&amp;key="+key))
	})
	mux.HandleFunc("/get.php", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") == "K1" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, "epub bytes")
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	// Act
	book, err := libgen.NewPlusMirror(server.URL).Download(context.Background(), testMD5)

	// Assert
	require.NoError(t, err)
	defer book.Stream.Close()
	content, err := io.ReadAll(book.Stream)
	require.NoError(t, err)
	assert.Equal(t, "epub bytes", string(content))
	assert.Equal(t, int32(2), adsCalls.Load())
}

func TestIsCoverURL_URLs_AcceptsOnlyLibgenImages(t *testing.T) {
	t.Parallel()

	cases := []struct {
		url  string
		want bool
	}{
		{"https://libgen.la/fictioncovers/10039000/2bb4_small.jpg", true},
		{"https://libgen.bz/covers/8374000/d98e_small.JPG", true},
		{"http://libgen.is/covers/1/a.png", true},
		{"https://libgen.la/ads.php?md5=x", false},
		{"https://libgen.la/img/blank.png", false},
		{"https://evil.com/covers/a.jpg", false},
		{"https://libgen.la.evil.com/a.jpg", false},
		{"https://user@libgen.la/a.jpg", false},
		{"file:///etc/a.jpg", false},
		{"", false},
	}

	for _, tc := range cases {
		t.Run(tc.url, func(t *testing.T) {
			t.Parallel()

			// Act
			got := libgen.IsCoverURL(tc.url)

			// Assert
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestLibgenCoverSource_Cover_NotALibgenImage_ReturnsErrInvalidCoverURL(t *testing.T) {
	t.Parallel()

	// Act
	cover, err := libgen.NewLibgenCoverSource().Cover(context.Background(), "http://127.0.0.1/covers/a.jpg")

	// Assert
	assert.ErrorIs(t, err, models.ErrInvalidCoverURL)
	assert.Nil(t, cover)
}

func plusPage(page string) string {
	row := `<tr><td></td><td><a href="edition.php?id=1">Book ` + page + `</a></td><td>Brandon Sanderson</td><td></td><td></td>` +
		`<td>Spanish</td><td>0</td><td></td><td>epub</td><td><a href="/ads.php?md5=0123456789abcdef0123456789abcdef">x</a></td></tr>`
	return `<html><body><table id="tablelibgen"><tr><th>h</th></tr>` + row + `</table>` +
		`<script type="text/javascript">paginator_example_top = new Paginator("paginator_example_top", 6, 25, ` + page + `, "/index.php?page=" );</script>` +
		`</body></html>`
}

func TestPlusMirror_Search_ManyPages_ReadsSeveralAndReturnsNext(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		page  int
		books []string
		next  int
	}{
		{"First pages", 1, []string{"Book 1", "Book 2", "Book 3", "Book 4"}, 5},
		{"Last pages", 5, []string{"Book 5", "Book 6"}, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				fmt.Fprint(w, plusPage(r.URL.Query().Get("page")))
			}))
			t.Cleanup(server.Close)

			// Act
			result, err := libgen.NewPlusMirror(server.URL).Search(context.Background(), "sanderson", models.SearchFieldAuthor, tc.page)

			// Assert
			require.NoError(t, err)
			titles := make([]string, len(result.Books))
			for i, b := range result.Books {
				titles[i] = b.Title
			}
			assert.Equal(t, tc.books, titles)
			assert.Equal(t, tc.next, result.NextPage)
			assert.Equal(t, int32(len(tc.books)), requests.Load())
		})
	}
}

func TestPlusMirror_Search_LaterPageFails_ReturnsItAsNext(t *testing.T) {
	t.Parallel()

	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		if page == "3" {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		fmt.Fprint(w, plusPage(page))
	}))
	t.Cleanup(server.Close)

	// Act
	result, err := libgen.NewPlusMirror(server.URL).Search(context.Background(), "sanderson", models.SearchFieldAuthor, 1)

	// Assert
	require.NoError(t, err)
	require.Len(t, result.Books, 2)
	assert.Equal(t, 3, result.NextPage)
}

func TestClassicMirror_Search_FullPage_ReturnsNextPage(t *testing.T) {
	t.Parallel()

	// Arrange
	var gotPage string
	rows := `<tr><td>ID</td></tr>`
	for i := 1; i <= 100; i++ {
		rows += fmt.Sprintf(`<tr><td>%d</td></tr>`, i)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/search.php", func(w http.ResponseWriter, r *http.Request) {
		gotPage = r.URL.Query().Get("page")
		fmt.Fprint(w, `<table class="c">`+rows+`</table>`)
	})
	mux.HandleFunc("/json.php", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"title":"Dune","extension":"epub"}]`)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	// Act
	result, err := libgen.NewClassicMirror(server.URL).Search(context.Background(), "dune", models.SearchFieldTitle, 2)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "2", gotPage)
	assert.Equal(t, 3, result.NextPage)
}
