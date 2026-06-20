package notifications

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type FirebasePushNotifier struct {
	client *messaging.Client
	ctx *context.Context
}

func (f *FirebasePushNotifier) Init() error {
	ctx := context.Background()
	opt := option.WithAuthCredentialsFile(option.AuthorizedUser, "")

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil { return err }

	client, err := app.Messaging(ctx)
	if err != nil { return err }
	
	f.client = client
	f.ctx = &ctx

	return nil
}

func (f *FirebasePushNotifier) Send(token, title, message string) error {
	if token == "" { return nil }
	
	firebaseMessage := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body: message,
		},
		Token: token,
	}

	_, err := f.client.Send(*
		f.ctx, firebaseMessage)
	return err
}