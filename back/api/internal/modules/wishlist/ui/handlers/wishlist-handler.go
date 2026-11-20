package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/ui/dtos"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/ui/requtil"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/additem"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/listitems"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/lookup"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/removeitem"
	"github.com/ismelen/inkomi/back/api/internal/platform/httpauth"
)

type WishlistHandler struct {
	addUC    *additem.AddItemUC
	removeUC *removeitem.RemoveItemUC
	listUC   *listitems.ListItemsUC
	lookupUC *lookup.LookupUC
}

func NewWishlistHandler(addUC *additem.AddItemUC, removeUC *removeitem.RemoveItemUC, listUC *listitems.ListItemsUC, lookupUC *lookup.LookupUC) *WishlistHandler {
	return &WishlistHandler{addUC: addUC, removeUC: removeUC, listUC: listUC, lookupUC: lookupUC}
}

// HandleList serves GET /wishlist.
func (h *WishlistHandler) HandleList(r *http.Request) (*[]dtos.ItemResponse, error) {
	items, err := h.listUC.Execute(r.Context(), userId(r))
	if err != nil {
		return nil, err
	}
	response := make([]dtos.ItemResponse, len(items))
	for i, item := range items {
		response[i] = dtos.NewItemResponse(item)
	}
	return &response, nil
}

// HandleAdd serves POST /wishlist.
func (h *WishlistHandler) HandleAdd(r *http.Request) (*dtos.ItemResponse, error) {
	body, err := requtil.DecodeJSON[dtos.AddItemRequest](r)
	if err != nil {
		return nil, err
	}
	item, err := h.addUC.Execute(r.Context(), userId(r), models.ItemKind(body.Kind), body.Title, body.Author, body.WorkKey, body.AuthorKey)
	if err != nil {
		return nil, mapError(err)
	}
	response := dtos.NewItemResponse(*item)
	return &response, nil
}

// HandleRemove serves DELETE /wishlist/{id}.
func (h *WishlistHandler) HandleRemove(r *http.Request) (*struct{}, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return nil, requtil.NewError(http.StatusBadRequest, "invalid id")
	}
	return nil, mapError(h.removeUC.Execute(r.Context(), userId(r), id))
}

// HandleLookup serves GET /wishlist/lookup?kind=book|author&q=<text>.
func (h *WishlistHandler) HandleLookup(r *http.Request) (*[]dtos.SuggestionResponse, error) {
	params := r.URL.Query()
	suggestions, err := h.lookupUC.Execute(r.Context(), models.ItemKind(params.Get("kind")), params.Get("q"))
	if err != nil {
		return nil, mapError(err)
	}
	response := make([]dtos.SuggestionResponse, len(suggestions))
	for i, s := range suggestions {
		response[i] = dtos.NewSuggestionResponse(s)
	}
	return &response, nil
}

func userId(r *http.Request) int {
	claims, _ := httpauth.GetUserClaims(r.Context())
	return claims.Id
}

func mapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, models.ErrInvalidKind), errors.Is(err, models.ErrEmptyTitle),
		errors.Is(err, models.ErrTooLong), errors.Is(err, models.ErrEmptyQuery):
		return requtil.NewError(http.StatusBadRequest, err.Error())
	case errors.Is(err, models.ErrDuplicate):
		return requtil.NewError(http.StatusConflict, err.Error())
	case errors.Is(err, models.ErrItemNotFound):
		return requtil.NewError(http.StatusNotFound, err.Error())
	case errors.Is(err, models.ErrWishlistFull):
		return requtil.NewError(http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, models.ErrLookupFailure):
		return requtil.NewError(http.StatusBadGateway, models.ErrLookupFailure.Error())
	}
	return err
}
