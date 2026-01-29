package documentBuilder

import "ismelen/ermc/internal/domain"

type BuilderI interface {
	Copy() BuilderI
	SetSettings(settings *domain.Settings) BuilderI
	Start(volume *domain.Volume) BuilderI
	Build() (string, error)
	AddPage(page *domain.Page) BuilderI
}
