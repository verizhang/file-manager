package s3

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/verizhang/file-manager/internal/errs"
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
		return nil, fmt.Errorf("%w: %v", errs.ErrGeneratePresignedURL, err)
	}

	return &storage.GeneratePresignedUploadURLResult{
		URL: request.URL,
	}, nil
}

func (s *Storage) HeadObject(
	ctx context.Context,
	bucket string,
	objectKey string,
) error {

	_, err := s.client.HeadObject(
		ctx,
		&s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(objectKey),
		},
	)

	if err != nil {

		var notFound *types.NotFound

		if errors.As(err, &notFound) {
			return errs.ErrFileNotUploaded
		}

		return err
	}

	return nil
}
