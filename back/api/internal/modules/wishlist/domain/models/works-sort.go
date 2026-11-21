package models

type WorksSort string

const (
	WorksSortPopular WorksSort = "popular"
	WorksSortNew     WorksSort = "new"
	WorksSortOld     WorksSort = "old"
	WorksSortTitle   WorksSort = "title"
)

func (s WorksSort) IsValid() bool {
	switch s {
	case WorksSortPopular, WorksSortNew, WorksSortOld, WorksSortTitle:
		return true
	}
	return false
}
