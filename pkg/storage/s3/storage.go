package s3

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/verizhang/file-manager/pkg/errs"
	"github.com/verizhang/file-manager/pkg/storage"
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

func (s *Storage) GeneratePresignedDownloadURL(
	ctx context.Context,
	opts storage.GeneratePresignedDownloadURLOptions,
) (*storage.GeneratePresignedDownloadURLResult, error) {
	request, err := s.presigner.PresignGetObject(
		ctx,
		&awss3.GetObjectInput{
			Bucket: &opts.Bucket,
			Key:    &opts.ObjectKey,
		},
		func(po *awss3.PresignOptions) {
			po.Expires = opts.Expiry
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrGeneratePresignedURL, err)
	}

	return &storage.GeneratePresignedDownloadURLResult{
		URL: request.URL,
	}, nil
}

func (s *Storage) DeleteObject(
	ctx context.Context,
	bucket string,
	objectKey string,
) (*storage.DeleteObjectResult, error) {
	_, err := s.client.DeleteObject(
		ctx,
		&awss3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(objectKey),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrDeleteFile, err)
	}

	return &storage.DeleteObjectResult{}, nil
}

func (s *Storage) GetObject(
	ctx context.Context,
	opts storage.GetObjectOptions,
) (*storage.GetObjectResult, error) {
	output, err := s.client.GetObject(
		ctx,
		&awss3.GetObjectInput{
			Bucket: aws.String(opts.Bucket),
			Key:    aws.String(opts.ObjectKey),
		},
	)
	if err != nil {
		return nil, err
	}

	return &storage.GetObjectResult{
		Body: output.Body,
	}, nil
}

func (s *Storage) CreateMultipartUpload(
	ctx context.Context,
	opts storage.CreateMultipartUploadOptions,
) (*storage.CreateMultipartUploadResult, error) {
	output, err := s.client.CreateMultipartUpload(
		ctx,
		&awss3.CreateMultipartUploadInput{
			Bucket:      &opts.Bucket,
			Key:         &opts.ObjectKey,
			ContentType: &opts.ContentType,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrCreateMultipartUpload, err)
	}

	return &storage.CreateMultipartUploadResult{
		UploadID: *output.UploadId,
	}, nil
}

func (s *Storage) GeneratePresignedMultipartUploadURL(
	ctx context.Context,
	opts storage.GeneratePresignedMultipartUploadURLOptions,
) (*storage.GeneratePresignedMultipartUploadURLResult, error) {
	request, err := s.presigner.PresignUploadPart(
		ctx,
		&awss3.UploadPartInput{
			Bucket:     &opts.Bucket,
			Key:        &opts.ObjectKey,
			UploadId:   &opts.UploadID,
			PartNumber: aws.Int32(opts.PartNumber),
		},
		func(po *awss3.PresignOptions) {
			po.Expires = opts.Expiry
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrGeneratePresignedURL, err)
	}

	return &storage.GeneratePresignedMultipartUploadURLResult{
		URL:     request.URL,
		Headers: convertHTTPHeaders(request.SignedHeader),
	}, nil
}

func (s *Storage) CompleteMultipartUpload(
	ctx context.Context,
	opts storage.CompleteMultipartUploadOptions,
) (*storage.CompleteMultipartUploadResult, error) {
	completedParts := make([]types.CompletedPart, len(opts.Parts))
	for i, part := range opts.Parts {
		completedParts[i] = types.CompletedPart{
			PartNumber: aws.Int32(part.PartNumber),
			ETag:       aws.String(part.ETag),
		}
	}

	output, err := s.client.CompleteMultipartUpload(
		ctx,
		&awss3.CompleteMultipartUploadInput{
			Bucket:   &opts.Bucket,
			Key:      &opts.ObjectKey,
			UploadId: &opts.UploadID,
			MultipartUpload: &types.CompletedMultipartUpload{
				Parts: completedParts,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrCompleteMultipartUpload, err)
	}

	return &storage.CompleteMultipartUploadResult{
		ETag: *output.ETag,
	}, nil
}

func (s *Storage) AbortMultipartUpload(
	ctx context.Context,
	opts storage.AbortMultipartUploadOptions,
) (*storage.AbortMultipartUploadResult, error) {
	_, err := s.client.AbortMultipartUpload(
		ctx,
		&awss3.AbortMultipartUploadInput{
			Bucket:   &opts.Bucket,
			Key:      &opts.ObjectKey,
			UploadId: &opts.UploadID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrAbortMultipartUpload, err)
	}

	return &storage.AbortMultipartUploadResult{}, nil
}

func convertHTTPHeaders(httpHeaders map[string][]string) map[string]string {
	headers := make(map[string]string)
	for key, values := range httpHeaders {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}
