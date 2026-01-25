package manga

type ConverterRequestDTO struct {
	Author         string `form:"author"`
	Title          string `form:"title"`
	Profile        string `form:"profile"`
	Merge          bool   `form:"merge"`
	FirstVolumeNum int    `form:"firstVolumeNum"`
	Format         string `form:"format"`
}
