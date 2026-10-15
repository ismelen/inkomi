package models_test

import (
	"testing"

	"github.com/ismelen/inkomi/back/library-service/internal/domain/models"
	"github.com/stretchr/testify/assert"
)

func TestNewMD5_Value_ValidatesAndLowercases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		value string
		want  models.MD5
		err   error
	}{
		{"Lowercase", "0123456789abcdef0123456789abcdef", "0123456789abcdef0123456789abcdef", nil},
		{"Uppercase with spaces", " 0123456789ABCDEF0123456789ABCDEF ", "0123456789abcdef0123456789abcdef", nil},
		{"Too short", "0123456789abcdef", "", models.ErrInvalidMD5},
		{"Not hex", "z123456789abcdef0123456789abcdef", "", models.ErrInvalidMD5},
		{"Filename", "book.epub", "", models.ErrInvalidMD5},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			md5, err := models.NewMD5(tc.value)

			// Assert
			assert.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.want, md5)
		})
	}
}

func TestBookDownload_Filename_TitleAndExt_ReturnsSafeName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		title string
		ext   string
		want  string
	}{
		{"Plain", "Dune", "epub", "Dune.epub"},
		{"Uppercase ext with dot", "Dune", ".EPUB", "Dune.epub"},
		{"Default ext", "Dune", "", "Dune.epub"},
		{"Path separators", "AC/DC: a \\ story", "pdf", "AC-DC: a - story.pdf"},
		{"Surrounding spaces", "  Dune  ", "mobi", "Dune.mobi"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			book := &models.BookDownload{Title: tc.title, Ext: tc.ext}

			// Act
			filename, err := book.Filename()

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tc.want, filename)
		})
	}
}
