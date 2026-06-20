package cloud

import (
	"context"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type GoogleCloud struct {
	folder  string
	service *drive.Service
}

func New(accessToken, folder string) (*GoogleCloud, error) {
	token := &oauth2.Token{AccessToken: accessToken}
	ctx := context.Background()

	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))

	service, err := drive.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil { return nil, err }

	return &GoogleCloud{
		folder: folder,
		service: service,
	}, nil
}

func (gc *GoogleCloud) Upload(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	filename := filepath.Base(path)

	f := &drive.File{
		Name:    filename,
		Parents: []string{gc.folder},
	}

	_, err = gc.service.Files.Create(f).Media(file).Do()
	return "", err
}
