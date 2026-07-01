package dto

// TransactionConfigRequest is the HTTP form input for a conversion request.
type TransactionConfigRequest struct {
	Author      string `form:"author"`
	Title       string `form:"title"`
	Profile     string `form:"profile"`
	Merge       bool   `form:"merge"`
	Cloud       bool   `form:"cloud"`
	CloudToken  string `form:"cloud_token"`
	CloudFolder string `form:"cloud_folder"`
	NotifyToken string `form:"notify_token"`
	Md5s        string `form:"md5s"`
}
