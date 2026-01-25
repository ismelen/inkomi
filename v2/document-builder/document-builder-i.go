package documentBuilder

type BuilderI interface {
	GetOutput() (string, error)
}
