package GoogleCloud

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/option"
)

func UploadToDrive(clientToken, folderID, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	ctx := context.Background()

	token := &oauth2.Token{
		AccessToken: clientToken,
		TokenType:   "Bearer",
	}
	conf := &oauth2.Config{}
	client := conf.Client(ctx, token)

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return err
	}

	driveFile := &drive.File{
		Title:   filepath.Base(filePath),
		Parents: []*drive.ParentReference{{Id: folderID}},
	}

	_, err = srv.Files.Insert(driveFile).
		Media(file).
		Do()

	if err != nil {
		return fmt.Errorf("unable to upload file: %v", err)
	}

	return nil
}
