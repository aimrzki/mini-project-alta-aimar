package helper

import (
	"bytes"
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"io"
)

func UploadImageToGCS(imageData []byte, imageName string) (string, error) {
	ctx := context.Background()

	client, err := storage.NewClient(ctx, option.WithCredentialsFile("key/credentials.json"))
	if err != nil {
		return "", err
	}

	bucketName := "relaverse"

	object := client.Bucket(bucketName).Object(imageName)
	wc := object.NewWriter(ctx)

	wc.ContentType = "image/jpeg"

	if _, err := io.Copy(wc, bytes.NewReader(imageData)); err != nil {
		wc.Close()
		return "", err
	}

	if err := wc.Close(); err != nil {
		return "", err
	}

	imageURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, imageName)
	return imageURL, nil
}
