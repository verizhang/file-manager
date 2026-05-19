package grpc

import (
	"context"
	filev1 "github.com/verizhang/file-manager/gen/go/file/v1"
	"github.com/verizhang/file-manager/internal/service"
)

type FileHandler struct {
	filev1.UnimplementedFileServiceServer

	fileService service.FileService
}

func NewFileHandler(
	fileService service.FileService,
) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

func (h *FileHandler) CreateMultipartUpload(
	ctx context.Context,
	req *filev1.CreateMultipartUploadRequest,
) (*filev1.CreateMultipartUploadResponse, error) {

	return h.fileService.CreateMultipartUpload(
		ctx,
		req,
	)
}