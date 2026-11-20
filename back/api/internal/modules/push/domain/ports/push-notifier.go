package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/push/domain/models"
)

type IPushNotifier interface {
	Send(ctx context.Context, token string, message models.PushMessage) error
}
