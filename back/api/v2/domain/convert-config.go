package domain

type ConvertConfig struct {
	Author    string `form:"author"`
	Title     string `form:"title"`
	Profile   string `form:"profile"`
	Merge     bool   `form:"merge"`
	Id        string
	Cloud     bool   `form:"cloud"`
	UserToken string `form:"token"`
}