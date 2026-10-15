package joinepubs_test

import (
	"testing"

	"github.com/ismelen/inkomi/back/contracts/jobs"
	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/epub-worker/internal/usecases/joinepubs"
	"github.com/stretchr/testify/assert"
)

func sizedItem(title string, size int64) models.JoinItem {
	return models.JoinItem{Msg: jobs.Msg{Id: title, Title: title}, Size: size}
}

func titles(blocks [][]models.JoinItem) [][]string {
	out := make([][]string, len(blocks))
	for i, block := range blocks {
		for _, it := range block {
			out[i] = append(out[i], it.Msg.Title)
		}
	}
	return out
}

func TestSplitBlocks(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		items    []models.JoinItem
		expected [][]string
	}{
		{"Empty", nil, [][]string{}},
		{"All fit", []models.JoinItem{sizedItem("a", 40), sizedItem("b", 60)}, [][]string{{"a", "b"}}},
		{"Exact boundary stays", []models.JoinItem{sizedItem("a", 50), sizedItem("b", 50), sizedItem("c", 1)}, [][]string{{"a", "b"}, {"c"}}},
		{"Never exceeds", []models.JoinItem{sizedItem("a", 60), sizedItem("b", 60), sizedItem("c", 30)}, [][]string{{"a"}, {"b", "c"}}},
		{"Giant item alone", []models.JoinItem{sizedItem("a", 10), sizedItem("b", 500), sizedItem("c", 10)}, [][]string{{"a"}, {"b"}, {"c"}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			blocks := joinepubs.SplitBlocks(tc.items, 100)

			// Assert
			assert.Equal(t, tc.expected, titles(blocks))
		})
	}
}
