package SharedInterfaces

type Notifier struct {
	observers []Observer
}

func (this *Notifier) Register(o Observer) {
	this.observers = append(this.observers, o)
}

func (this *Notifier) Notify(event any) {
	for _, o := range this.observers {
		o.OnNotify(event)
	}
}
