package service

import (
	"context"
	filev1 "github.com/verizhang/file-manager/gen/go/file/v1"
	"github.com/verizhang/file-manager/internal/repository"
	"github.com/verizhang/file-manager/internal/storage"
)

type fileService struct {
	fileRepository repository.FileRepository
	minioStorage  *storage.Minio
}

func NewFileService(
	fileRepository repository.FileRepository,
	minioStorage *storage.Minio,
) FileService {
	return &fileService{
		fileRepository: fileRepository,
		minioStorage:  minioStorage,
	}
}

func (s *fileService) CreateMultipartUpload(
	ctx context.Context,
	req *filev1.CreateMultipartUploadRequest,
) (*filev1.CreateMultipartUploadResponse, error) {

	// TODO:
	// implement multipart upload initiation

	return &filev1.CreateMultipartUploadResponse{}, nil
}