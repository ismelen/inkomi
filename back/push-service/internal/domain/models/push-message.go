package models

import (
	"fmt"
	"strconv"

	"github.com/ismelen/inkomi/back/common/jobs"
)

const PushTypeSent = "sent"

type PushMessage struct {
	Title string
	Body  string
	Data  map[string]string
}

func NewSentMessage(msg jobs.Msg) PushMessage {
	name := msg.Title
	if name == "" {
		name = msg.Filename
	}

	body := fmt.Sprintf("%s is ready", name)
	if msg.ToCloud {
		body = fmt.Sprintf("%s was sent to Dropbox", name)
	}

	data := map[string]string{
		"type":     PushTypeSent,
		"id":       msg.Id,
		"filename": msg.Filename,
		"toCloud":  strconv.FormatBool(msg.ToCloud),
	}
	if msg.Result != nil {
		data["parentId"] = msg.Result.ParentId
	}

	return PushMessage{Title: "Book ready", Body: body, Data: data}
}
