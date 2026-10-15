package push

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/ismelen/inkomi/back/push-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/push-service/internal/domain/ports"
	"google.golang.org/api/option"
)

var _ ports.IPushNotifier = (*FirebasePushNotifier)(nil)

type FirebasePushNotifier struct {
	client *messaging.Client
}

// NewFirebasePushNotifier authenticates with the service account JSON at credentialsFile.
func NewFirebasePushNotifier(ctx context.Context, credentialsFile string) (*FirebasePushNotifier, error) {
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("init firebase app: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("init firebase messaging: %w", err)
	}
	return &FirebasePushNotifier{client: client}, nil
}

func (f *FirebasePushNotifier) Send(ctx context.Context, token string, message models.PushMessage) error {
	_, err := f.client.Send(ctx, &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: message.Title,
			Body:  message.Body,
		},
		Data:    message.Data,
		Android: &messaging.AndroidConfig{Priority: "high"},
	})
	if err != nil {
		return fmt.Errorf("send fcm message: %w", err)
	}
	return nil
}
