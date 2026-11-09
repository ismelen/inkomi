package handlers

import (
	"net/http"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/ports"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/httpauth"
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
