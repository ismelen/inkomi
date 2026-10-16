package dtos

import (
	"time"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
)

// SourceRowResponse is a source row without its owner and config, the same fields the websocket sends.
type SourceRowResponse struct {
	Id          string              `json:"id"`
	Title       string              `json:"title"`
	Filename    string              `json:"filename"`
	Kepubify    bool                `json:"kepubify"`
	ToCloud     bool                `json:"toCloud"`
	IsRTL       bool                `json:"isRtl"`
	Join        bool                `json:"join"`
	Status      models.SourceStatus `json:"status"`
	ErrorMsg    *string             `json:"errorMsg"`
	Size        *int64              `json:"size"`
	IsFile      bool                `json:"isFile"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
	CompletedAt *time.Time          `json:"completedAt"`
	ParentId    *string             `json:"parentId"`
	Position    int                 `json:"position"`
	IsResult    bool                `json:"isResult"`
}

func NewSourceRowResponse(s *models.Source) SourceRowResponse {
	return SourceRowResponse{
		Id:          s.Id,
		Title:       s.Title,
		Filename:    s.Filename,
		Kepubify:    s.Kepubify,
		ToCloud:     s.ToCloud,
		IsRTL:       s.IsRTL,
		Join:        s.Join,
		Status:      s.Status,
		ErrorMsg:    s.ErrorMsg,
		Size:        s.Size,
		IsFile:      s.IsFile,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		CompletedAt: s.CompletedAt,
		ParentId:    s.ParentId,
		Position:    s.Position,
		IsResult:    s.IsResult,
	}
}
