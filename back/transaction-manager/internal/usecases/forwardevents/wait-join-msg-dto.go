package forwardevents

type WaitJoinMsgDTO struct {
	Id       string `json:"id"`
	UserId   int    `json:"userId"`
	FolderId string `json:"folderId"`
}
