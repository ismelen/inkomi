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

func (gc *GoogleCloud) Init(accessToken, folder string) error {
	token := &oauth2.Token{AccessToken: accessToken}
	ctx := context.Background()

	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))

	service, err := drive.NewService(ctx, option.WithHTTPClient(httpClient))

	gc.folder = folder
	gc.service = service

	return err
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
