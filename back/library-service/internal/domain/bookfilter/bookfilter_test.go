package bookfilter_test

import (
	"testing"

	"github.com/ismelen/inkomi/back/library-service/internal/domain/bookfilter"
	"github.com/ismelen/inkomi/back/library-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/library-service/internal/shared/filter"
	"github.com/stretchr/testify/assert"
)

func titles(books []models.Book) []string {
	out := make([]string, 0, len(books))
	for _, b := range books {
		out = append(out, b.Title)
	}
	return out
}

func TestFilterChain_Filter_Books_KeepsMatching(t *testing.T) {
	t.Parallel()

	books := []models.Book{
		{Title: "Dune", Language: "English", Extension: "epub"},
		{Title: "Duna", Language: "Spanish", Extension: "epub"},
		{Title: "Dune Messiah", Language: "English", Extension: "pdf"},
		{Title: "Children of Dune", Language: "english", Extension: "EPUB"},
	}

	cases := []struct {
		name     string
		language string
		formats  []string
		want     []string
	}{
		{"No filters", "", nil, []string{"Dune", "Duna", "Dune Messiah", "Children of Dune"}},
		{"Language", "English", nil, []string{"Dune", "Dune Messiah", "Children of Dune"}},
		{"Format", "", []string{"epub"}, []string{"Dune", "Duna", "Children of Dune"}},
		{"Both", "english", []string{"pdf"}, []string{"Dune Messiah"}},
		{"Nothing matches", "French", nil, []string{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			chain := filter.Use(
				&bookfilter.LanguageFilter{Language: tc.language},
				&bookfilter.FormatFilter{Formats: tc.formats},
				&bookfilter.DeduplicateFilter{},
			)

			// Act
			ok, result := chain.Filter(books)

			// Assert
			assert.True(t, ok)
			assert.Equal(t, tc.want, titles(result))
		})
	}
}

func TestDeduplicateFilter_Filter_RepeatedAndSpam_KeepsOnePerTitle(t *testing.T) {
	t.Parallel()

	// Arrange
	books := []models.Book{
		{Title: "Dune"},
		{Title: "dune", MD5: "abc"},
		{Title: "Dúne", MD5: "def"},
		{Title: "978-3-16-148410-0; 978-0"},
		{Title: "Dune epub gratis"},
		{Title: "Hi"},
		{Title: "Foundation", MD5: "ghi"},
	}

	// Act
	ok, result := (&bookfilter.DeduplicateFilter{}).Filter(books)

	// Assert
	assert.True(t, ok)
	assert.Equal(t, []models.Book{
		{Title: "dune", MD5: "abc"},
		{Title: "Foundation", MD5: "ghi"},
	}, result)
}
