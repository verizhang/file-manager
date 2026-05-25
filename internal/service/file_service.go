package service

import (
	"context"

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
