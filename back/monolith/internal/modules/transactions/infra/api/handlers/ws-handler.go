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

// HandleConnect upgrades the connection: the upgrader writes the response, also on failure, so
// it isn't wrapped like the JSON handlers.
func (h *WSHandler) HandleConnect(w http.ResponseWriter, r *http.Request) {
	claims, _ := httpauth.GetUserClaims(r.Context())
	// Register logs the failure and the upgrader already answered the client
	_ = h.sockethub.Register(claims.Id, w, r)
}
