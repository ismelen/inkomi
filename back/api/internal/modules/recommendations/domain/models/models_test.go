package models_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
)

func TestEmbedding_Cosine_Vectors_ComparesDirection(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		a, b models.Embedding
		want float64
	}{
		{"Same", models.Embedding{1, 0}, models.Embedding{2, 0}, 1},
		{"Orthogonal", models.Embedding{1, 0}, models.Embedding{0, 1}, 0},
		{"Opposite", models.Embedding{1, 0}, models.Embedding{-1, 0}, -1},
		{"Other size", models.Embedding{1, 0}, models.Embedding{1}, 0},
		{"Zero", models.Embedding{0, 0}, models.Embedding{1, 0}, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act & Assert
			assert.InDelta(t, tc.want, tc.a.Cosine(tc.b), 1e-9)
		})
	}
}

func TestWeightedMean_Vectors_NormalizesTheWeightedSum(t *testing.T) {
	t.Parallel()

	// Act
	mean := models.WeightedMean([]models.Embedding{{1, 0}, {0, 1}}, []float64{3, 1})

	// Assert
	assert.InDelta(t, 0.9487, mean[0], 1e-4)
	assert.InDelta(t, 0.3162, mean[1], 1e-4)
	assert.Nil(t, models.WeightedMean(nil, nil))
	assert.Nil(t, models.WeightedMean([]models.Embedding{{1, 0}}, []float64{1, 2}))
}

func TestWork_Passage_Work_JoinsTextAndCapsIt(t *testing.T) {
	t.Parallel()

	// Arrange
	subjects := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}
	work := models.Work{Title: " La nena ", Authors: []string{"Carmen Mola"}, Subjects: subjects, Description: strings.Repeat("x", 700)}

	// Act
	passage := work.Passage()

	// Assert
	assert.True(t, strings.HasPrefix(passage, "La nena · Carmen Mola · a, b, c, d, e, f, g, h, i, j · "))
	assert.NotContains(t, passage, "k")
	assert.Equal(t, strings.Repeat("x", 600), passage[strings.LastIndex(passage, " · ")+len(" · "):])
	assert.Equal(t, "Dune", (&models.Work{Title: "Dune"}).Passage())
}

func TestCleanTitle_Names_RemovesNoise(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"La novia gitana.epub":                   "La novia gitana",
		"Carmen_Mola_-_La_nena (z-lib.org).epub": "Carmen Mola - La nena",
		"Dune [Herbert, Frank].kepub.epub":       "Dune",
		"  Los   huérfanos ":                     "Los huérfanos",
		"Un libro.PDF":                           "Un libro",
	}
	for input, want := range cases {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			// Act & Assert
			assert.Equal(t, want, models.CleanTitle(input))
		})
	}
}

func TestBudget_Take_Requests_SpendsUntilEmpty(t *testing.T) {
	t.Parallel()

	// Arrange
	budget := models.NewBudget(3)

	// Act & Assert
	assert.True(t, budget.Take(2))
	assert.False(t, budget.Take(2))
	assert.True(t, budget.Take(1))
	assert.Equal(t, 0, budget.Left())
}

func TestSection_IsValid_Values_AcceptsTheThree(t *testing.T) {
	t.Parallel()

	for _, section := range models.Sections() {
		assert.True(t, section.IsValid())
	}
	assert.False(t, models.Section("popular").IsValid())
}
