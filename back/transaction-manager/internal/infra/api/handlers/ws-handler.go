package handlers

import (
	"net/http"

	"github.com/ismelen/inkomi/back/common/httpauth"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
)

type WSHandler struct {
	sockethub ports.ISocketHub
}

func NewWSHandler(sockethub ports.ISocketHub) *WSHandler {
	return &WSHandler{
		sockethub: sockethub,
	}
}

func (h *WSHandler) HandleConnect(w http.ResponseWriter, r *http.Request) (*any, error) {
	token, _ := httpauth.GetUserClaims(r.Context())
	err := h.sockethub.Register(token.Id, w, r)

	return nil, err
}
