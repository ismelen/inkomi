package SharedInterfaces

type Observer interface {
	OnNotify(event any)
}
