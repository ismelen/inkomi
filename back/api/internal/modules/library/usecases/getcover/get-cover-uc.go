package getcover

import (
	"context"
	"strings"

	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/ports"
)

// GetCoverUC serves the cover images of the search results through the api: the clients may
// not reach the libgen mirrors (blocked in some countries) while the server does.
type GetCoverUC struct {
	covers ports.ICoverSource
}

func NewGetCoverUC(covers ports.ICoverSource) *GetCoverUC {
	return &GetCoverUC{covers: covers}
}

func (g *GetCoverUC) Execute(ctx context.Context, url string) (*models.Cover, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return nil, models.ErrInvalidCoverURL
	}
	return g.covers.Cover(ctx, url)
}
