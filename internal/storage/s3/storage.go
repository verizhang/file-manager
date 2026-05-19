package s3

import (
	"context"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/verizhang/file-manager/internal/storage"
)

type Storage struct {
	client    *awss3.Client
	presigner *awss3.PresignClient
}

func NewStorage(client *awss3.Client) *Storage {
	return &Storage{
		client:    client,
		presigner: awss3.NewPresignClient(client),
	}
}

func (s *Storage) GeneratePresignedUploadURL(
	ctx context.Context,
	opts storage.GeneratePresignedUploadURLOptions,
) (*storage.GeneratePresignedUploadURLResult, error) {

	request, err := s.presigner.PresignPutObject(
		ctx,
		&awss3.PutObjectInput{
			Bucket:      &opts.Bucket,
			Key:         &opts.ObjectKey,
			ContentType: &opts.ContentType,
		},
		func(po *awss3.PresignOptions) {
			po.Expires = opts.Expiry
		},
	)
	if err != nil {
		return nil, err
	}

	return &storage.GeneratePresignedUploadURLResult{
		URL: request.URL,
	}, nil
}
