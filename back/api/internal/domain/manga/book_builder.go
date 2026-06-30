package manga

// BookBuilder is the port implemented by infra/epub to assemble an epub file.
type BookBuilder interface {
	SetSettings(settings *ImageSettings, profile *Profile) BookBuilder
	Start(name, outDir string) BookBuilder
	AddPage(page *Page, fstPage bool) BookBuilder
	Build() (string, error)
}
