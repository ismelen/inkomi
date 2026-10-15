package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/push-service/internal/domain/models"
)

type IPushNotifier interface {
	Send(ctx context.Context, token string, message models.PushMessage) error
}
