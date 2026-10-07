package usecases

import (
	"context"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
)

type UploadCompleteUC struct {
	sourceRepo ports.SourceRepository
}

func (u *UploadCompleteUC) Execute(ctx context.Context, userId int, sourceId string) error {
	src, err := u.sourceRepo.GetByIDAndUserID(ctx, sourceId, userId)
}
