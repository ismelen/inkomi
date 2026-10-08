package forwardevents

type ErrorMsgDTO struct {
	Id     string `json:"id"`
	UserId int    `json:"userId"`
	Error  string `json:"error"`
}
