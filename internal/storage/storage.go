package storage

import (
	"context"
	"time"
)

type GeneratePresignedUploadURLOptions struct {
	Bucket      string
	ObjectKey   string
	ContentType string
	Expiry      time.Duration
}

type GeneratePresignedUploadURLResult struct {
	URL string
}

type Storage interface {
	GeneratePresignedUploadURL(
		ctx context.Context,
		opts GeneratePresignedUploadURLOptions,
	) (*GeneratePresignedUploadURLResult, error)
	HeadObject(
		ctx context.Context,
		bucket string,
		objectKey string,
	) error
}
