package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	keyToHash   map[string]string
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
		keyToHash:   make(map[string]string),
	}

	if err := n.saveConfigs(operation); err != nil {
		return nil, err
	}

	if err := n.saveSources(operation); err != nil {
		return nil, err
	}

	for _, src := range data.Sources {
		src.Url, err = n.cloudStorage.GetUrl(src.Id, filepath.Ext(src.Filename))
	}

	return operation.data.Sources, nil
}

func (n *NewUploadRequestUC) saveSources(op *uploadOperation) error {
	return n.txManager.ExecuteTx(op.ctx, func(txCtx context.Context) (err error) {
		for i, s := range op.data.Sources {
			data := n.getSource(&s.NoItemsSourceDTO, op.userId)
			data.ShouldJoin = s.ShouldJoin
			if data.Kepubify, err = n.shouldKepubify(s.ConfigHash, op); err != nil {
				return err
			}
			if s.ConfigHash != nil {
				hash := op.keyToHash[*s.ConfigHash]
				data.ConfigHash = &hash
			}

			if s.Id, err = n.sourceRepo.Create(txCtx, data); err != nil {
				return err
			}
			s.Kepubify = data.Kepubify
			op.data.Sources[i] = s

			for j, item := range s.Items {
				data = n.getSource(&item, op.userId)
				data.FolderId = &s.Id
				if data.Kepubify, err = n.shouldKepubify(item.ConfigHash, op); err != nil {
					return err
				}
				if item.ConfigHash != nil {
					hash := op.keyToHash[*item.ConfigHash]
					data.ConfigHash = &hash
				}
				data.ShouldJoin = s.ShouldJoin
				item.Id, err = n.sourceRepo.Create(txCtx, data)
				if err != nil {
					return err
				}
				item.Kepubify = data.Kepubify
				op.data.Sources[i].Items[j] = item
			}
		}

		return nil
	})
}

func (n *NewUploadRequestUC) shouldKepubify(configKey *string, op *uploadOperation) (bool, error) {
	if configKey == nil {
		return op.mainEreader.IsKepub, nil
	}

	config, ok := op.data.Configs[*configKey]
	if !ok {
		return false, fmt.Errorf("invalid config reference")
	}

	if config.EReaderKey == op.mainEreader.Key {
		return op.mainEreader.IsKepub, nil
	}

	ereader, err := models.NewEreader(config.EReaderKey)
	if err != nil {
		return false, err
	}

	return ereader.IsKepub, nil
}

func (n *NewUploadRequestUC) getSource(s *dtos.NoItemsSourceDTO, userId int) *models.Source {
	return &models.Source{
		UserId:           userId,
		Title:            s.Title,
		Size:             s.Size,
		Filename:         s.Filename,
		Type:             s.Type,
		ReadingDirection: s.ReadingDirection,
	}
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

			op.keyToHash[key] = hash
		}
		return nil
	})

	return err
}
