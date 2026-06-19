package domain

type ConvertConfig struct {
	Author      string `form:"author"`
	Title       string `form:"title"`
	Profile     string `form:"profile"`
	Merge       bool   `form:"merge"`
	Id          string
	Cloud       bool   `form:"cloud"`
	CloudToken  string `form:"cloud_token,omitempty"`
	CloudFolder string `form:"cloud_folder,omitempty"`
	NotifyToken string `form:"notify_token,omitempty"`
}
