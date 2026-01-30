package cloud

type CloudI interface {
	Init(accesToken, folder string) error
	Upload(path string) error
}
