package service

import (
	"context"

	filev1 "github.com/verizhang/file-manager/gen/go/file/v1"
)

type FileService interface {
	CreateMultipartUpload(
		ctx context.Context,
		req *filev1.CreateMultipartUploadRequest,
	) (*filev1.CreateMultipartUploadResponse, error)
}