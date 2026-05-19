package service

import (
	"context"

	filev1 "github.com/verizhang/file-manager/gen/go/file/v1"
)

type FileService interface {
	CreateUploadURL(
		ctx context.Context,
		req *filev1.CreateUploadUrlRequest,
	) (*filev1.CreateUploadUrlResponse, error)
}
