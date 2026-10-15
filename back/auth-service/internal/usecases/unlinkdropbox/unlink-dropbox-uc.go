package unlinkdropbox

import (
	"context"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
)

type UnlinkDropboxUC struct {
	links ports.IDropboxLinkRepository
}

func NewUnlinkDropboxUC(links ports.IDropboxLinkRepository) *UnlinkDropboxUC {
	return &UnlinkDropboxUC{links: links}
}

// Execute stops uploading the files of the user to Dropbox; signing in with it still works.
func (u *UnlinkDropboxUC) Execute(ctx context.Context, userId int) error {
	return u.links.Delete(ctx, userId)
}
