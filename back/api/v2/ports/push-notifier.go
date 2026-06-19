package ports

type PushNotifier interface {
	Init() error
	Send(token, title, message string) error
}