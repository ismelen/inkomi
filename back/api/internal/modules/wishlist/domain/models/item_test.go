package models_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
)

func TestNewItem_Inputs_ValidatesAndTrims(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		kind  models.ItemKind
		title string
		err   error
	}{
		{"Book", models.ItemKindBook, "  La nena ", nil},
		{"Invalid kind", models.ItemKind("movie"), "x", models.ErrInvalidKind},
		{"Empty title", models.ItemKindBook, "   ", models.ErrEmptyTitle},
		{"Too long", models.ItemKindBook, strings.Repeat("a", 301), models.ErrTooLong},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			item, err := models.NewItem(1, tc.kind, tc.title, " Carmen Mola ", "", "")

			// Assert
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "La nena", item.Title)
			assert.Equal(t, "Carmen Mola", item.Author)
		})
	}
}

func TestItem_DedupKey_Kinds_UsesKeyOrNormalizedText(t *testing.T) {
	t.Parallel()

	// Arrange
	keyed, err := models.NewItem(1, models.ItemKindBook, "La nena", "", "OL1W", "")
	require.NoError(t, err)
	free, err := models.NewItem(1, models.ItemKindBook, "La Néna", "Carmen  Mola", "", "")
	require.NoError(t, err)
	author, err := models.NewItem(1, models.ItemKindAuthor, "Carmen Mola", "ignored", "OL1W", "OL9A")
	require.NoError(t, err)

	// Act & Assert
	assert.Equal(t, "ol:OL1W", keyed.DedupKey())
	assert.Equal(t, "la-nena|carmen--mola", free.DedupKey())
	assert.Equal(t, "ol:OL9A", author.DedupKey())
	assert.Empty(t, author.Author, "an author has no author")
	assert.Empty(t, author.WorkKey)
}
