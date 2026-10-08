package usecases_test

import (
	"context"
	"testing"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/dtos"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/usecases"
	"github.com/stretchr/testify/assert"
)

func TestNewUploadRequestUC_Execute_InvalidEreaderKey_ReturnsError(t *testing.T) {
	// Arrange
	uc := usecases.NewNewUploadRequestUC(
		&MockTxManager{},
		&MockSourceRepository{},
		&MockConfigRepository{},
		&MockCloudStorage{},
	)

	req := dtos.UploadRequestDTO{
		EReaderKey: "INVALID_KEY",
	}

	// Act
	_, err := uc.Execute(context.Background(), req, 1)

	// Assert
	assert.Error(t, err, "expected error for invalid ereader key")
}

func TestNewUploadRequestUC_Execute_ValidRequest_ReturnsSources(t *testing.T) {
	// Arrange
	txManager := &MockTxManager{}
	sourceRepo := &MockSourceRepository{}
	configRepo := &MockConfigRepository{}
	cloudStorage := &MockCloudStorage{
		GetUrlFn: func(id string, ext string) (string, error) {
			return "http://example.com/" + id + ext, nil
		},
	}

	uc := usecases.NewNewUploadRequestUC(
		txManager,
		sourceRepo,
		configRepo,
		cloudStorage,
	)

	req := dtos.UploadRequestDTO{
		EReaderKey: "K1",
		Sources: []dtos.SourceDTO{
			{
				NoItemsSourceDTO: dtos.NoItemsSourceDTO{
					Title:    "Source 1",
					Filename: "test1.zip",
					Type:     models.SourceTypeFile,
				},
			},
		},
	}

	// Act
	sources, err := uc.Execute(context.Background(), req, 1)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, sources, 1)
	assert.Equal(t, "http://example.com/.zip", sources[0].Url) // ID is empty because mock returns empty
}

func TestNewUploadRequestUC_Execute_WithConfigs_SavesConfigsAndSources(t *testing.T) {
	// Arrange
	txManager := &MockTxManager{}
	configHashCreated := ""
	configRepo := &MockConfigRepository{
		CreateFn: func(ctx context.Context, config *models.Config) (string, error) {
			configHashCreated = config.Hash
			return config.Hash, nil
		},
	}
	sourceRepo := &MockSourceRepository{}
	cloudStorage := &MockCloudStorage{}

	uc := usecases.NewNewUploadRequestUC(
		txManager,
		sourceRepo,
		configRepo,
		cloudStorage,
	)

	confKey := "conf1"
	req := dtos.UploadRequestDTO{
		EReaderKey: "K1",
		Configs: map[string]dtos.ConfigDTO{
			confKey: {
				EReaderKey: "K1",
			},
		},
		Sources: []dtos.SourceDTO{
			{
				NoItemsSourceDTO: dtos.NoItemsSourceDTO{
					Title:     "Source Config",
					Filename:  "test2.zip",
					Type:      models.SourceTypeFile,
					ConfigKey: &confKey,
				},
			},
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), req, 1)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, configHashCreated, "expected config to be saved and hash generated")
}

func TestNewUploadRequestUC_Execute_SourceWithItemsNotFolder_ReturnsError(t *testing.T) {
	// Arrange
	uc := usecases.NewNewUploadRequestUC(
		&MockTxManager{},
		&MockSourceRepository{},
		&MockConfigRepository{},
		&MockCloudStorage{},
	)

	req := dtos.UploadRequestDTO{
		EReaderKey: "K1",
		Sources: []dtos.SourceDTO{
			{
				NoItemsSourceDTO: dtos.NoItemsSourceDTO{
					Type: models.SourceTypeFile,
				},
				Items: []dtos.NoItemsSourceDTO{
					{Type: models.SourceTypeFile},
				},
			},
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), req, 1)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source with items must be of type folder")
}

func TestNewUploadRequestUC_Execute_ChildInheritsConfigHash_UsesParentHash(t *testing.T) {
	// Arrange
	txManager := &MockTxManager{}
	configRepo := &MockConfigRepository{
		CreateFn: func(ctx context.Context, config *models.Config) (string, error) {
			return config.Hash, nil
		},
	}

	var capturedChild *models.Source
	sourceRepo := &MockSourceRepository{
		CreateFn: func(ctx context.Context, source *models.Source) (string, error) {
			if source.FolderId != nil {
				capturedChild = source
			}
			return "some-id", nil
		},
	}
	cloudStorage := &MockCloudStorage{}

	uc := usecases.NewNewUploadRequestUC(
		txManager,
		sourceRepo,
		configRepo,
		cloudStorage,
	)

	parentConfKey := "parent-conf"
	req := dtos.UploadRequestDTO{
		EReaderKey: "K1",
		Configs: map[string]dtos.ConfigDTO{
			parentConfKey: {EReaderKey: "K1"},
		},
		Sources: []dtos.SourceDTO{
			{
				NoItemsSourceDTO: dtos.NoItemsSourceDTO{
					Type:      models.SourceTypeFolder,
					ConfigKey: &parentConfKey,
				},
				Items: []dtos.NoItemsSourceDTO{
					{
						Type:     models.SourceTypeFile,
						Filename: "child.zip",
					},
				},
			},
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), req, 1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, capturedChild)
	assert.NotNil(t, capturedChild.ConfigHash)
}

func TestNewUploadRequestUC_Execute_ChildInheritsParentAttributes_MatchesParent(t *testing.T) {
	// Arrange
	txManager := &MockTxManager{}
	configRepo := &MockConfigRepository{}

	var capturedChild *models.Source
	sourceRepo := &MockSourceRepository{
		CreateFn: func(ctx context.Context, source *models.Source) (string, error) {
			if source.FolderId != nil {
				capturedChild = source
			}
			return "parent-id", nil
		},
	}
	cloudStorage := &MockCloudStorage{}

	uc := usecases.NewNewUploadRequestUC(
		txManager,
		sourceRepo,
		configRepo,
		cloudStorage,
	)

	req := dtos.UploadRequestDTO{
		EReaderKey: "KoMT",
		Sources: []dtos.SourceDTO{
			{
				ShouldJoin: true,
				NoItemsSourceDTO: dtos.NoItemsSourceDTO{
					Type:             models.SourceTypeFolder,
					ReadingDirection: models.ReadingDirectionRTL,
				},
				Items: []dtos.NoItemsSourceDTO{
					{
						Type: models.SourceTypeFile,
					},
				},
			},
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), req, 1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, capturedChild)
	assert.Equal(t, "parent-id", *capturedChild.FolderId)
	assert.True(t, capturedChild.ShouldJoin)
	assert.Equal(t, models.ReadingDirectionRTL, capturedChild.ReadingDirection)
	assert.True(t, capturedChild.Kepubify)
}

func TestNewUploadRequestUC_Execute_FolderAndLibrary_EmptyUrl(t *testing.T) {
	// Arrange
	uc := usecases.NewNewUploadRequestUC(
		&MockTxManager{},
		&MockSourceRepository{},
		&MockConfigRepository{},
		&MockCloudStorage{
			GetUrlFn: func(id string, ext string) (string, error) {
				return "should-not-be-called", nil
			},
		},
	)

	req := dtos.UploadRequestDTO{
		EReaderKey: "K1",
		Sources: []dtos.SourceDTO{
			{
				NoItemsSourceDTO: dtos.NoItemsSourceDTO{
					Type: models.SourceTypeFolder,
				},
				Items: []dtos.NoItemsSourceDTO{
					{
						Type: models.SourceTypeLibrary,
					},
				},
			},
		},
	}

	// Act
	sources, err := uc.Execute(context.Background(), req, 1)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, sources, 1)
	assert.Empty(t, sources[0].Url)
	assert.Len(t, sources[0].Items, 1)
	assert.Empty(t, sources[0].Items[0].Url)
}
