package dtos

import (
	"time"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
)

type UploadListResponse struct {
	Uploads    []UploadResponse `json:"uploads"`
	NextCursor *string          `json:"nextCursor"`
	SyncedAt   time.Time        `json:"syncedAt"`
}

type UploadResponse struct {
	SourceRowResponse
	Children []SourceRowResponse `json:"children"`
}

func NewUploadListResponse(page *models.SourcePage) *UploadListResponse {
	uploads := make([]UploadResponse, len(page.Trees))
	for i := range page.Trees {
		tree := &page.Trees[i]
		children := make([]SourceRowResponse, len(tree.Children))
		for j := range tree.Children {
			children[j] = NewSourceRowResponse(&tree.Children[j])
		}
		uploads[i] = UploadResponse{
			SourceRowResponse: NewSourceRowResponse(&tree.Source),
			Children:          children,
		}
	}

	res := &UploadListResponse{Uploads: uploads, SyncedAt: page.SyncedAt.UTC()}
	if page.NextCursor != nil {
		cursor := page.NextCursor.Encode()
		res.NextCursor = &cursor
	}

	return res
}
