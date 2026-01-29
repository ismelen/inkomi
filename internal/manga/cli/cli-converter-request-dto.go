package manga

type CliConverterRequestDTO struct {
	Author         string
	Title          string
	Profile        string
	Merge          bool
	FirstVolumeNum int
	Format         string
	InputDir       string
	RamLimit       int
}
