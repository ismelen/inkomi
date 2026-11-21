package models

type WorksPage struct {
	Works   []Suggestion
	Total   int
	Page    int
	HasMore bool
}
