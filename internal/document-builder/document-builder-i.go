package documentBuilder

import "ismelen/ermc/internal/domain"

type BuilderI interface {
	SetSettings(settings *domain.Settings) BuilderI
	Start(name string) BuilderI
	Build() (string, error)
	AddPage(page *domain.Page, fstPage bool) BuilderI
}
