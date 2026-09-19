package helper

import (
	"context"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var cld *cloudinary.Cloudinary

type UploadFileResult struct {
	SecureURL string
	PublicID  string
}

func InitCloudinary(cloudName, apiKey, apiSecret string) error {
	var err error
	cld, err = cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	return err
}

func UploadImage(file multipart.File, folder string) (string, error) {
	ctx := context.Background()

	uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: folder,
	})

	if err != nil {
		return "", err
	}

	return uploadResult.SecureURL, nil
}

func UploadFile(file multipart.File, folder string, fileName string) (*UploadFileResult, error) {
	ctx := context.Background()

	uploadResult, err := cld.Upload.Upload(
		ctx,
		file,
		uploader.UploadParams{
			Folder:   folder,
			PublicID: fileName,
		},
	)

	if err != nil {
		return nil, err
	}

	return &UploadFileResult{
		SecureURL: uploadResult.SecureURL,
		PublicID:  uploadResult.PublicID,
	}, nil
}
