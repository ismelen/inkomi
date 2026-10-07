package handlers

import (
	"net/http"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/middlewares"
)

type WSHandler struct {
	sockethub ports.SocketHub
}

func NewWSHandler(sockethub ports.SocketHub) *WSHandler {
	return &WSHandler{
		sockethub: sockethub,
	}
}

func (h *WSHandler) HandleConnect(w http.ResponseWriter, r *http.Request) (*any, error) {
	token, _ := middlewares.GetUserClaims(r.Context())
	err := h.sockethub.Register(token.Id, w, r)

	return nil, err
}
