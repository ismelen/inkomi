package convert

// PushNotifier is the port implemented by infra/push to send push notifications.
type PushNotifier interface {
	Init() error
	Send(token, title, message string) error
}
