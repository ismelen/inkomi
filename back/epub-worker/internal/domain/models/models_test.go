package models_test

import (
	"testing"

	"github.com/ismelen/inkomi/back/contracts/jobs"
	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/models"
	"github.com/stretchr/testify/assert"
)

func TestExtractChapterNumber(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		expected float64
		ok       bool
	}{
		{"One Piece Ch. 12", 12, true},
		{"Capítulo 3.5", 3.5, true},
		{"012---title", 12, true},
		{"Episode 7 (2021)", 7, true},
		{"Vol 2 - 45", 45, true},
		{"Prologue", 0, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			n, ok := models.ExtractChapterNumber(tc.name)

			// Assert
			assert.Equal(t, tc.ok, ok)
			assert.Equal(t, tc.expected, n)
		})
	}
}

func item(title string, size int64) models.JoinItem {
	return models.JoinItem{Msg: jobs.Msg{Id: title, Title: title}, Size: size}
}

func TestBlockTitle(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		block    []models.JoinItem
		index    int
		total    int
		expected string
	}{
		{"Range", []models.JoinItem{item("Ch. 1", 0), item("Ch. 12", 0)}, 0, 2, "Manga Ch.[1 - 12]"},
		{"Single chapter", []models.JoinItem{item("Ch. 13", 0)}, 1, 2, "Manga Ch.13"},
		{"No numbers, one block", []models.JoinItem{item("Prologue", 0)}, 0, 1, "Manga"},
		{"No numbers, several blocks", []models.JoinItem{item("Prologue", 0)}, 1, 3, "Manga Pt.2"},
		{
			"Falls back to filename",
			[]models.JoinItem{{Msg: jobs.Msg{Title: "Start", Filename: "chapter 4.epub"}}},
			0, 1, "Manga Ch.4",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			title := models.BlockTitle("Manga", tc.block, tc.index, tc.total)

			// Assert
			assert.Equal(t, tc.expected, title)
		})
	}
}
