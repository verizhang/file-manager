package service

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/verizhang/file-manager/internal/config"
	"github.com/verizhang/file-manager/internal/errs"
	"github.com/verizhang/file-manager/internal/model"
	"github.com/verizhang/file-manager/internal/repository"
	"github.com/verizhang/file-manager/internal/storage"
)

type fileService struct {
	cfg            *config.Config
	logger         *zap.Logger
	storage        storage.Storage
	fileRepository repository.FileRepository
}

func NewFileService(
	cfg *config.Config,
	logger *zap.Logger,
	storage storage.Storage,
	fileRepository repository.FileRepository,
) FileService {
	return &fileService{
		cfg:            cfg,
		logger:         logger,
		storage:        storage,
		fileRepository: fileRepository,
	}
}

func (s *fileService) CreateUploadUrl(
	ctx context.Context,
	req *CreateUploadRequest,
) (*CreateUploadResponse, error) {

	fileID := uuid.NewString()

	objectKey := GenerateObjectKey(
		fileID,
		req.FileName,
	)

	uploadURL, err := s.storage.GeneratePresignedUploadURL(
		ctx,
		storage.GeneratePresignedUploadURLOptions{
			Bucket:      s.cfg.S3.Bucket,
			ObjectKey:   objectKey,
			ContentType: req.ContentType,
		},
	)
	if err != nil {
		return nil, errs.ErrGeneratePresignedURL
	}

	file := &model.File{
		ID:          fileID,
		ObjectKey:   objectKey,
		Bucket:      s.cfg.S3.Bucket,
		FileName:    req.FileName,
		ContentType: req.ContentType,
		Size:        req.Size,
		Status:      model.FileStatusPending,
	}

	err = s.fileRepository.Create(
		ctx,
		file,
	)
	if err != nil {
		return nil, errs.ErrFileNotFound
	}

	return &CreateUploadResponse{
		File:      file,
		UploadURL: uploadURL.URL,
	}, nil
}

func (s *fileService) CompleteUpload(
	ctx context.Context,
	req *CompleteUploadRequest,
) (*CompleteUploadResponse, error) {

	file, err := s.fileRepository.GetByID(
		ctx,
		req.FileID,
	)
	if err != nil {
		return nil, err
	}

	if file.Status == model.FileStatusCompleted {
		return &CompleteUploadResponse{
			File: file,
		}, nil
	}

	err = s.storage.HeadObject(
		ctx,
		file.Bucket,
		file.ObjectKey,
	)
	if err != nil {
		return nil, err
	}

	err = s.fileRepository.UpdateStatus(
		ctx,
		file.ID,
		model.FileStatusCompleted,
	)
	if err != nil {
		return nil, err
	}

	file.Status = model.FileStatusCompleted

	return &CompleteUploadResponse{
		File: file,
	}, nil
}

func GenerateObjectKey(
	fileID string,
	fileName string,
) string {

	now := time.Now().UTC()

	ext := filepath.Ext(fileName)

	return fmt.Sprintf(
		"%d/%02d/%02d/%s%s",
		now.Year(),
		now.Month(),
		now.Day(),
		fileID,
		ext,
	)
}
