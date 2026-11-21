package getwork_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/getwork"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/mocks"
)

func TestGetWorkUC_Execute_Stored_ReturnsItWithoutCatalog(t *testing.T) {
	t.Parallel()

	// Arrange
	uc := getwork.NewGetWorkUC(&mocks.MockWorkRepository{
		GetFn: func(ctx context.Context, workKey string) (*models.WorkDetails, error) {
			return &models.WorkDetails{WorkKey: workKey, Description: "Stored."}, nil
		},
	}, &mocks.MockCatalogLookup{
		WorkFn: func(ctx context.Context, workKey string) (models.WorkDetails, error) {
			t.Fatal("the catalog isn't asked for a stored work")
			return models.WorkDetails{}, nil
		},
	}, &mocks.MockSynopsisRepository{}, &mocks.MockSynopsisLookup{})

	// Act
	work, err := uc.Execute(context.Background(), "OL5W", "")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Stored.", work.Description)
}

func TestGetWorkUC_Execute_NotStored_FetchesAndSaves(t *testing.T) {
	t.Parallel()

	// Arrange
	var saved models.WorkDetails
	uc := getwork.NewGetWorkUC(&mocks.MockWorkRepository{
		SaveFn: func(ctx context.Context, work models.WorkDetails) error {
			saved = work
			return nil
		},
	}, &mocks.MockCatalogLookup{
		WorkFn: func(ctx context.Context, workKey string) (models.WorkDetails, error) {
			return models.WorkDetails{Title: "Mistborn", Description: "Ash falls."}, nil
		},
	}, &mocks.MockSynopsisRepository{}, &mocks.MockSynopsisLookup{})

	// Act
	work, err := uc.Execute(context.Background(), "OL5W", "")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Ash falls.", work.Description)
	assert.Equal(t, models.WorkDetails{WorkKey: "OL5W", Title: "Mistborn", Description: "Ash falls."}, saved)
}

func TestGetWorkUC_Execute_Failures_MapsThem(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		workKey string
		catalog error
		err     error
	}{
		{"Invalid key", "../x", nil, models.ErrInvalidWorkKey},
		{"Author key", "OL1A", nil, models.ErrInvalidWorkKey},
		{"Unknown", "OL5W", models.ErrWorkNotFound, models.ErrWorkNotFound},
		{"Catalog down", "OL5W", errors.New("timeout"), models.ErrLookupFailure},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			uc := getwork.NewGetWorkUC(&mocks.MockWorkRepository{}, &mocks.MockCatalogLookup{
				WorkFn: func(ctx context.Context, workKey string) (models.WorkDetails, error) {
					return models.WorkDetails{}, tc.catalog
				},
			}, &mocks.MockSynopsisRepository{}, &mocks.MockSynopsisLookup{})

			// Act
			_, err := uc.Execute(context.Background(), tc.workKey, "")

			// Assert
			assert.ErrorIs(t, err, tc.err)
		})
	}
}

func storedWork() *mocks.MockWorkRepository {
	return &mocks.MockWorkRepository{
		GetFn: func(ctx context.Context, workKey string) (*models.WorkDetails, error) {
			return &models.WorkDetails{WorkKey: workKey, Title: "Mistborn", Description: "Ash falls."}, nil
		},
	}
}

func TestGetWorkUC_Execute_LanguageNotStored_LooksUpAndSavesTheSynopsis(t *testing.T) {
	t.Parallel()

	// Arrange
	var saved models.Synopsis
	uc := getwork.NewGetWorkUC(storedWork(), &mocks.MockCatalogLookup{}, &mocks.MockSynopsisRepository{
		SaveFn: func(ctx context.Context, synopsis models.Synopsis) error {
			saved = synopsis
			return nil
		},
	}, &mocks.MockSynopsisLookup{
		SynopsisFn: func(ctx context.Context, workKey, lang string) (string, error) {
			return "Cae ceniza.", nil
		},
	})

	// Act
	work, err := uc.Execute(context.Background(), "OL5W", "es")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Cae ceniza.", work.Description)
	assert.Equal(t, models.Synopsis{WorkKey: "OL5W", Language: "es", Text: "Cae ceniza."}, saved)
}

func TestGetWorkUC_Execute_SynopsisStored_UsesItWithoutLookup(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		text      string
		fetchedAt time.Time
		want      string
	}{
		{"Found", "Cae ceniza.", time.Now().Add(-365 * 24 * time.Hour), "Cae ceniza."},
		{"Recently missing", "", time.Now().Add(-time.Hour), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			uc := getwork.NewGetWorkUC(storedWork(), &mocks.MockCatalogLookup{}, &mocks.MockSynopsisRepository{
				GetFn: func(ctx context.Context, workKey, lang string) (*models.Synopsis, error) {
					return &models.Synopsis{WorkKey: workKey, Language: lang, Text: tc.text, FetchedAt: tc.fetchedAt}, nil
				},
			}, &mocks.MockSynopsisLookup{
				SynopsisFn: func(ctx context.Context, workKey, lang string) (string, error) {
					t.Fatal("a stored synopsis isn't looked up")
					return "", nil
				},
			})

			// Act
			work, err := uc.Execute(context.Background(), "OL5W", "es")

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.want, work.Description)
		})
	}
}

func TestGetWorkUC_Execute_MissingLongAgo_LooksItUpAgain(t *testing.T) {
	t.Parallel()

	// Arrange
	uc := getwork.NewGetWorkUC(storedWork(), &mocks.MockCatalogLookup{}, &mocks.MockSynopsisRepository{
		GetFn: func(ctx context.Context, workKey, lang string) (*models.Synopsis, error) {
			return &models.Synopsis{Text: "", FetchedAt: time.Now().Add(-60 * 24 * time.Hour)}, nil
		},
	}, &mocks.MockSynopsisLookup{
		SynopsisFn: func(ctx context.Context, workKey, lang string) (string, error) {
			return "Cae ceniza.", nil
		},
	})

	// Act
	work, err := uc.Execute(context.Background(), "OL5W", "es")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Cae ceniza.", work.Description)
}

func TestGetWorkUC_Execute_NoSynopsisInLanguage_FallsBackOnlyForEnglish(t *testing.T) {
	t.Parallel()

	cases := []struct {
		lang string
		want string
	}{
		{"es", ""},
		{"en", "Ash falls."},
	}
	for _, tc := range cases {
		t.Run(tc.lang, func(t *testing.T) {
			t.Parallel()

			// Arrange
			uc := getwork.NewGetWorkUC(storedWork(), &mocks.MockCatalogLookup{}, &mocks.MockSynopsisRepository{}, &mocks.MockSynopsisLookup{})

			// Act
			work, err := uc.Execute(context.Background(), "OL5W", tc.lang)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.want, work.Description)
		})
	}
}

func TestGetWorkUC_Execute_LookupFails_AnswersWithoutSynopsisAndDoesNotSave(t *testing.T) {
	t.Parallel()

	// Arrange
	uc := getwork.NewGetWorkUC(storedWork(), &mocks.MockCatalogLookup{}, &mocks.MockSynopsisRepository{
		SaveFn: func(ctx context.Context, synopsis models.Synopsis) error {
			t.Fatal("a failed lookup isn't saved")
			return nil
		},
	}, &mocks.MockSynopsisLookup{
		SynopsisFn: func(ctx context.Context, workKey, lang string) (string, error) {
			return "", errors.New("quota exceeded")
		},
	})

	// Act
	work, err := uc.Execute(context.Background(), "OL5W", "es")

	// Assert
	require.NoError(t, err)
	assert.Empty(t, work.Description)
	assert.Equal(t, "Mistborn", work.Title)
}

func TestGetWorkUC_Execute_InvalidLanguage_Rejects(t *testing.T) {
	t.Parallel()

	// Arrange
	uc := getwork.NewGetWorkUC(storedWork(), &mocks.MockCatalogLookup{}, &mocks.MockSynopsisRepository{}, &mocks.MockSynopsisLookup{})

	// Act
	_, err := uc.Execute(context.Background(), "OL5W", "spanish")

	// Assert
	assert.ErrorIs(t, err, models.ErrInvalidLanguage)
}
