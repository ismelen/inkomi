package newupload

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/ports"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/api/dtos"
)

type NewUploadRequestUC struct {
	txManager    ports.ITxManager
	sourceRepo   ports.ISourceRepository
	configRepo   ports.IConfigRepository
	cloudStorage ports.ICloudStorage
	quota        ports.IStorageQuota
	estimate     models.StorageEstimateOptions
}

func NewNewUploadRequestUC(
	txManager ports.ITxManager,
	sourceRepo ports.ISourceRepository,
	configRepo ports.IConfigRepository,
	cloudStorage ports.ICloudStorage,
	quota ports.IStorageQuota,
	estimate models.StorageEstimateOptions,
) *NewUploadRequestUC {
	return &NewUploadRequestUC{
		txManager:    txManager,
		sourceRepo:   sourceRepo,
		configRepo:   configRepo,
		cloudStorage: cloudStorage,
		quota:        quota,
		estimate:     estimate,
	}
}

type uploadOperation struct {
	userId      int
	data        *dtos.UploadRequestDTO
	ctx         context.Context
	mainEreader *models.EReader
}

func (n *NewUploadRequestUC) Execute(ctx context.Context, data dtos.UploadRequestDTO, userId int) ([]dtos.SourceDTO, error) {
	ereader, err := models.NewEreader(data.EReaderKey)
	if err != nil {
		return nil, err
	}

	operation := &uploadOperation{
		userId:      userId,
		data:        &data,
		ctx:         ctx,
		mainEreader: ereader,
	}

	if err := n.saveConfigs(operation); err != nil {
		return nil, err
	}
	// the space check and the transaction that records it must not interleave with another upload
	err = n.quota.Exclusive(func() error {
		return n.txManager.ExecuteTx(operation.ctx, func(txCtx context.Context) error {
			if err := n.saveSources(txCtx, operation); err != nil {
				return err
			}
			return n.quota.Reserve(txCtx, n.reservations(userId, data.Sources))
		})
	})

	return data.Sources, err
}

// reservations estimates the disk peak of every root of the request.
func (n *NewUploadRequestUC) reservations(userId int, sources []dtos.SourceDTO) []models.StorageReservation {
	out := make([]models.StorageReservation, 0, len(sources))
	for _, root := range sources {
		req := models.StorageRequest{Join: root.ShouldJoin}
		files := root.Items
		if root.Type != models.SourceTypeFolder {
			files = []dtos.NoItemsSourceDTO{root.NoItemsSourceDTO}
		}
		for _, f := range files {
			item := models.StorageItem{
				IsLibrary: f.Type == models.SourceTypeLibrary,
				IsEpub:    strings.EqualFold(filepath.Ext(f.Filename), ".epub"),
			}
			if f.Size != nil {
				item.Size = *f.Size
			}
			req.Items = append(req.Items, item)
			req.Kepub = req.Kepub || f.Kepubify
		}
		out = append(out, models.StorageReservation{
			RootId: root.Id,
			UserId: userId,
			Bytes:  models.EstimateStoragePeak(req, n.estimate),
		})
	}
	return out
}

func (n *NewUploadRequestUC) saveSources(txCtx context.Context, operation *uploadOperation) error {
	data := operation.data
	for i := range data.Sources {
		topSrc := &data.Sources[i]
		if len(topSrc.Items) > 0 && topSrc.Type != models.SourceTypeFolder {
			topSrc.Type = models.SourceTypeFolder
		}
		if err := n.saveSource(txCtx, operation, &topSrc.NoItemsSourceDTO, topSrc.ShouldJoin, nil, 0); err != nil {
			return err
		}

		for j := range topSrc.Items {
			if err := n.saveSource(txCtx, operation, &topSrc.Items[j], topSrc.ShouldJoin, topSrc, j); err != nil {
				return err
			}
		}
	}
	return nil
}

func (n *NewUploadRequestUC) saveSource(ctx context.Context, op *uploadOperation, src *dtos.NoItemsSourceDTO, join bool, parent *dtos.SourceDTO, position int) error {
	if src.Type == models.SourceTypeFolder {
		src.Size = new(int64)
	}

	var parentId, configHash *string

	if src.ConfigKey != nil {
		if config, ok := op.data.Configs[*src.ConfigKey]; ok {
			configHash = &config.Hash
			src.Kepubify = config.Kepubify
		}
	}

	if parent != nil {
		src.Kepubify = parent.Kepubify
		src.ReadingDirection = parent.ReadingDirection
		src.ToCloud = parent.ToCloud
		parentId = new(parent.Id)

		if configHash == nil && parent.ConfigKey != nil {
			if config, ok := op.data.Configs[*parent.ConfigKey]; ok {
				configHash = &config.Hash
				src.Kepubify = config.Kepubify
			}
		}
	}

	if configHash == nil {
		ereader, err := models.NewEreader(op.data.EReaderKey)
		if err != nil {
			return err
		}
		src.Kepubify = ereader.IsKepub
	}

	id, err := n.sourceRepo.Create(ctx, &models.Source{
		UserId:    op.userId,
		Title:     src.Title,
		Filename:  src.Filename,
		Kepubify:  src.Kepubify,
		ToCloud:   src.ToCloud,
		ConfigKey: configHash,
		IsRTL:     src.ReadingDirection == models.ReadingDirectionRTL,
		Join:      join,
		Size:      src.Size,
		IsFile:    src.Type != models.SourceTypeFolder,
		ParentId:  parentId,
		Position:  position,
	})
	if err != nil {
		return err
	}

	src.Id = id
	if src.Type == models.SourceTypeFile {
		var size int64
		if src.Size != nil {
			size = *src.Size
		}
		src.Url, err = n.cloudStorage.GetUrl(src.Id, src.Filename, size)
	}

	return err
}

func (n *NewUploadRequestUC) saveConfigs(op *uploadOperation) error {
	err := n.txManager.ExecuteTx(op.ctx, func(txCtx context.Context) error {
		for key, config := range op.data.Configs {
			bytes, err := json.Marshal(config)
			if err != nil {
				return err
			}

			hashBytes := sha256.Sum256(bytes)
			hash := hex.EncodeToString(hashBytes[:])

			if _, err := n.configRepo.Create(txCtx, &models.Config{
				Hash: hash,
				Data: string(bytes),
			}); err != nil {
				return err
			}

			config.Hash = hash
			op.data.Configs[key] = config
		}
		return nil
	})

	return err
}
