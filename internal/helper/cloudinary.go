package helper

import (
	"context"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var cld *cloudinary.Cloudinary

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
