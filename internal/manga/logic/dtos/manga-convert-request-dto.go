package MangaDtos

type MangaConvertRequestDTO struct {
	GoogleCloudFolder string `form:"googleCloudFolder" json:"googleCloudFolder"`
	// OutputFilename      string `form:"outputFilename" json:"outputFilename"`
	Author              string `form:"author" json:"author"`
	Profile             string `form:"profile" json:"profile"`
	Title               string `form:"title" json:"title"`
	Merge               bool   `form:"merge" json:"merge"`
	StartingVolumeCount int    `form:"startVolumeCount" json:"startVolumeCount"`
}
