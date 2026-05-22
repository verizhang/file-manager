package service

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	filev1 "github.com/verizhang/file-manager/gen/go/file/v1"
	"github.com/verizhang/file-manager/internal/config"
	"github.com/verizhang/file-manager/internal/model"
	"github.com/verizhang/file-manager/internal/repository"
	"github.com/verizhang/file-manager/internal/storage"

	"go.uber.org/zap"
)

type fileService struct {
	config         *config.Config
	logger         *zap.Logger
	storage        storage.Storage
	fileRepository repository.FileRepository
}

func NewFileService(
	config *config.Config,
	logger *zap.Logger,
	storage storage.Storage,
	fileRepository repository.FileRepository,
) FileService {
	return &fileService{
		storage:        storage,
		logger:         logger,
		config:         config,
		fileRepository: fileRepository,
	}
}

func (s *fileService) CreateUploadURL(
	ctx context.Context,
	req *filev1.CreateUploadUrlRequest,
) (*filev1.CreateUploadUrlResponse, error) {

	fileID := uuid.NewString()

	objectKey := generateObjectKey(
		fileID,
		req.GetFileName(),
	)

	file := &model.File{
		ID:          fileID,
		Bucket:      s.config.S3.Bucket,
		ObjectKey:   objectKey,
		FileName:    req.GetFileName(),
		ContentType: req.GetContentType(),
		Size:        req.GetSize(),
		Status:      model.FileStatusPending,
	}

	result, err := s.storage.GeneratePresignedUploadURL(
		ctx,
		storage.GeneratePresignedUploadURLOptions{
			Bucket:      s.config.S3.Bucket,
			ObjectKey:   objectKey,
			ContentType: req.GetContentType(),
			Expiry:      15 * time.Minute,
		},
	)
	if err != nil {
		s.logger.Error("failed generate presigned upload url",
			zap.Error(err),
		)

		return nil, err
	}

	err = s.fileRepository.Create(ctx, file)
	if err != nil {
		s.logger.Error("failed write file metadata to database", zap.Error(err))
		return nil, err
	}

	return &filev1.CreateUploadUrlResponse{
		FileId:    fileID,
		UploadUrl: result.URL,
		ObjectKey: objectKey,
	}, nil
}

func generateObjectKey(
	fileID string,
	fileName string,
) string {

	ext := filepath.Ext(fileName)

	date := time.Now().Format("2006/01/02")

	return fmt.Sprintf(
		"%s/%s%s",
		date,
		fileID,
		ext,
	)
}
