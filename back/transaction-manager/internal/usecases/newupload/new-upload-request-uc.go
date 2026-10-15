package newupload

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/dtos"
)

type NewUploadRequestUC struct {
	txManager    ports.ITxManager
	sourceRepo   ports.ISourceRepository
	configRepo   ports.IConfigRepository
	cloudStorage ports.ICloudStorage
}

func NewNewUploadRequestUC(
	txManager ports.ITxManager,
	sourceRepo ports.ISourceRepository,
	configRepo ports.IConfigRepository,
	cloudStorage ports.ICloudStorage,
) *NewUploadRequestUC {
	return &NewUploadRequestUC{
		txManager,
		sourceRepo,
		configRepo,
		cloudStorage,
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
	err = n.txManager.ExecuteTx(operation.ctx, func(txCtx context.Context) error {
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
	})

	return data.Sources, err
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
		src.Url, err = n.cloudStorage.GetUrl(src.Id, filepath.Ext(src.Filename))
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
