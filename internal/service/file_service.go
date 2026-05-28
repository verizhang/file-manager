package service

import (
	"context"
	"time"

	"github.com/verizhang/file-manager/internal/model"
)

type FileService interface {
	CreateUploadUrl(
		ctx context.Context,
		req *CreateUploadRequest,
	) (*CreateUploadResponse, error)

	CompleteUpload(
		ctx context.Context,
		req *CompleteUploadRequest,
	) (*CompleteUploadResponse, error)

	CreateMultipartUpload(
		ctx context.Context,
		req *CreateMultipartUploadRequest,
	) (*CreateMultipartUploadResponse, error)

	CreateMultipartUploadUrl(
		ctx context.Context,
		req *CreateMultipartUploadUrlRequest,
	) (*CreateMultipartUploadUrlResponse, error)

	CompleteMultipartUpload(
		ctx context.Context,
		req *CompleteMultipartUploadRequest,
	) (*CompleteMultipartUploadResponse, error)

	AbortMultipartUpload(
		ctx context.Context,
		req *AbortMultipartUploadRequest,
	) (*AbortMultipartUploadResponse, error)
}

type CreateUploadRequest struct {
	FileName    string
	ContentType string
	Size        int64
}

type CreateUploadResponse struct {
	File      *model.File
	UploadURL string
}

type CompleteUploadRequest struct {
	FileID string
}

type CompleteUploadResponse struct {
	File *model.File
}

// Multipart Upload
type CreateMultipartUploadRequest struct {
	FileName    string
	ContentType string
	Size        int64
}

type CreateMultipartUploadResponse struct {
	FileID    string
	UploadID  string
	ObjectKey string
	PartSize  int64
	TotalParts int32
	ExpiresAt time.Time
}

type CreateMultipartUploadUrlRequest struct {
	FileID     string
	UploadID   string
	ObjectKey  string
	PartNumber int32
}

type CreateMultipartUploadUrlResponse struct {
	UploadURL string
	Headers   map[string]string
}

type MultipartPart struct {
	PartNumber int32
	ETag       string
}

type CompleteMultipartUploadRequest struct {
	UploadID  string
	ObjectKey string
	Parts     []MultipartPart
}

type CompleteMultipartUploadResponse struct {
	File *model.File
}

type AbortMultipartUploadRequest struct {
	UploadID  string
	ObjectKey string
}

type AbortMultipartUploadResponse struct {
	// No specific fields needed for now, just a successful response
}
