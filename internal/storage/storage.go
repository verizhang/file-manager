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

// Multipart upload options and results
type CreateMultipartUploadOptions struct {
	Bucket      string
	ObjectKey   string
	ContentType string
}

type CreateMultipartUploadResult struct {
	UploadID string
}

type GeneratePresignedMultipartUploadURLOptions struct {
	Bucket      string
	ObjectKey   string
	UploadID    string
	PartNumber  int32
	ContentType string
	Expiry      time.Duration
}

type GeneratePresignedMultipartUploadURLResult struct {
	URL     string
	Headers map[string]string
}

type CompletedPart struct {
	PartNumber int32
	ETag       string
}

type CompleteMultipartUploadOptions struct {
	Bucket    string
	ObjectKey string
	UploadID  string
	Parts     []CompletedPart
}

type CompleteMultipartUploadResult struct {
	ETag string
}

type AbortMultipartUploadOptions struct {
	Bucket    string
	ObjectKey string
	UploadID  string
}

type AbortMultipartUploadResult struct {
	// No specific fields needed for now
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

	// Multipart Upload methods
	CreateMultipartUpload(
		ctx context.Context,
		opts CreateMultipartUploadOptions,
	) (*CreateMultipartUploadResult, error)
	GeneratePresignedMultipartUploadURL(
		ctx context.Context,
		opts GeneratePresignedMultipartUploadURLOptions,
	) (*GeneratePresignedMultipartUploadURLResult, error)
	CompleteMultipartUpload(
		ctx context.Context,
		opts CompleteMultipartUploadOptions,
	) (*CompleteMultipartUploadResult, error)
	AbortMultipartUpload(
		ctx context.Context,
		opts AbortMultipartUploadOptions,
	) (*AbortMultipartUploadResult, error)
}