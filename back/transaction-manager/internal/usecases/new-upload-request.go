package usecases

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
	txManager    ports.TxManager
	sourceRepo   ports.SourceRepository
	configRepo   ports.ConfigRepository
	cloudStorage ports.CloudStorage
}

func NewNewUploadRequestUC(
	txManager ports.TxManager,
	sourceRepo ports.SourceRepository,
	configRepo ports.ConfigRepository,
	cloudStorage ports.CloudStorage,
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

	for i, topSrc := range data.Sources {
		if len(topSrc.Items) > 0 && topSrc.Type != models.SourceTypeFolder {
			topSrc.Type = models.SourceTypeFolder
		}
		n.saveSource(operation, &topSrc.NoItemsSourceDTO, &topSrc)
		data.Sources[i] = topSrc

		for j, childSrc := range topSrc.Items {
			n.saveSource(operation, &childSrc, &topSrc)
			data.Sources[i].Items[j] = childSrc
		}
	}

	return data.Sources, nil
}

func (n *NewUploadRequestUC) saveSource(op *uploadOperation, src *dtos.NoItemsSourceDTO, parent *dtos.SourceDTO) error {
	if src.Type == models.SourceTypeFolder {
		src.Size = new(int64)
	}

	shouldJoin := false
	var folderId, configHash *string

	if src.ConfigKey != nil {
		if config, ok := op.data.Configs[*src.ConfigKey]; ok {
			configHash = &config.Hash
			src.Kepubify = config.Kepubify
		}
	}

	if parent != nil {
		src.Kepubify = parent.Kepubify
		shouldJoin = parent.ShouldJoin
		src.ReadingDirection = parent.ReadingDirection
		folderId = &parent.Id

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

	id, err := n.sourceRepo.Create(op.ctx, &models.Source{
		UserId:           op.userId,
		Size:             src.Size,
		Filename:         src.Filename,
		Title:            src.Title,
		Kepubify:         src.Kepubify,
		Type:             src.Type,
		ShouldJoin:       shouldJoin,
		ReadingDirection: src.ReadingDirection,
		FolderId:         folderId,
		ConfigHash:       configHash,
	})
	if err != nil {
		return err
	}

	src.Id = id
	if src.Type != models.SourceTypeFolder {
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
