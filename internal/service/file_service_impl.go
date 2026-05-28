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
			Expiry:      s.cfg.S3.PresignedConfig.UploadExpireMinutes,
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

func (s *fileService) CreateMultipartUpload(
	ctx context.Context,
	req *CreateMultipartUploadRequest,
) (*CreateMultipartUploadResponse, error) {
	fileID := uuid.NewString()

	objectKey := GenerateObjectKey(
		fileID,
		req.FileName,
	)

	minPartSize := s.cfg.Multipart.PartSize

	partSize := minPartSize
	if req.Size > 0 && req.Size < minPartSize {
		partSize = req.Size
	}

	totalParts := int32(1)
	if req.Size > 0 {
		totalParts = int32((req.Size + partSize - 1) / partSize)
	}

	createMultipartOutput, err := s.storage.CreateMultipartUpload(
		ctx,
		storage.CreateMultipartUploadOptions{
			Bucket:      s.cfg.S3.Bucket,
			ObjectKey:   objectKey,
			ContentType: req.ContentType,
		},
	)
	if err != nil {
		s.logger.Error("failed to create multipart upload", zap.Error(err))
		return nil, errs.ErrCreateMultipartUpload
	}

	file := &model.File{
		ID:          fileID,
		UploadID:    &createMultipartOutput.UploadID,
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
		s.logger.Error("failed to create file metadata for multipart upload", zap.Error(err))
		// Attempt to abort the multipart upload if metadata creation fails
		_, abortErr := s.storage.AbortMultipartUpload(ctx, storage.AbortMultipartUploadOptions{
			Bucket:    s.cfg.S3.Bucket,
			ObjectKey: objectKey,
			UploadID:  createMultipartOutput.UploadID,
		})
		if abortErr != nil {
			s.logger.Error("failed to abort multipart upload after metadata creation failed", zap.Error(abortErr))
		}
		return nil, errs.ErrCreateFileMetadata
	}

	return &CreateMultipartUploadResponse{
		FileID:     fileID,
		UploadID:   createMultipartOutput.UploadID,
		ObjectKey:  objectKey,
		PartSize:   partSize,
		TotalParts: totalParts,
	}, nil
}

func (s *fileService) CreateMultipartUploadUrl(
	ctx context.Context,
	req *CreateMultipartUploadUrlRequest,
) (*CreateMultipartUploadUrlResponse, error) {
	file, err := s.fileRepository.GetByID(ctx, req.FileID)
	if err != nil {
		return nil, err
	}

	if file.UploadID == nil || *file.UploadID != req.UploadID {
		return nil, errs.ErrMultipartUploadNotFound
	}

	if file.ObjectKey != req.ObjectKey {
		return nil, errs.ErrMultipartUploadNotFound
	}

	presignedURLOutput, err := s.storage.GeneratePresignedMultipartUploadURL(
		ctx,
		storage.GeneratePresignedMultipartUploadURLOptions{
			Bucket:      s.cfg.S3.Bucket,
			ObjectKey:   req.ObjectKey,
			UploadID:    req.UploadID,
			PartNumber:  req.PartNumber,
			ContentType: file.ContentType,
			Expiry:      s.cfg.S3.PresignedConfig.UploadExpireMinutes,
		},
	)
	if err != nil {
		s.logger.Error("failed to generate presigned multipart upload URL", zap.Error(err))
		return nil, errs.ErrGeneratePresignedURL
	}

	return &CreateMultipartUploadUrlResponse{
		UploadURL: presignedURLOutput.URL,
		Headers:   presignedURLOutput.Headers,
	}, nil
}

func (s *fileService) CompleteMultipartUpload(
	ctx context.Context,
	req *CompleteMultipartUploadRequest,
) (*CompleteMultipartUploadResponse, error) {
	file, err := s.fileRepository.GetByObjectKey(ctx, req.ObjectKey) // Assuming GetByObjectKey exists or needs to be created
	if err != nil {
		return nil, err
	}

	if file.UploadID == nil || *file.UploadID != req.UploadID {
		return nil, errs.ErrMultipartUploadNotFound
	}

	if file.Status == model.FileStatusCompleted {
		return &CompleteMultipartUploadResponse{
			File: file,
		}, nil
	}

	if len(req.Parts) == 0 {
		return nil, errs.ErrInvalidMultipartUploadParts
	}

	storageParts := make([]storage.CompletedPart, len(req.Parts))
	for i, p := range req.Parts {
		storageParts[i] = storage.CompletedPart{
			PartNumber: p.PartNumber,
			ETag:       p.ETag,
		}
	}

	completeMultipartOutput, err := s.storage.CompleteMultipartUpload(
		ctx,
		storage.CompleteMultipartUploadOptions{
			Bucket:    file.Bucket,
			ObjectKey: file.ObjectKey,
			UploadID:  req.UploadID,
			Parts:     storageParts,
		},
	)
	if err != nil {
		s.logger.Error("failed to complete multipart upload", zap.Error(err))
		return nil, errs.ErrCompleteMultipartUpload
	}

	// Update file status and ETag
	etag := completeMultipartOutput.ETag
	err = s.fileRepository.UpdateStatusAndETag(
		ctx,
		file.ID,
		model.FileStatusCompleted,
		&etag,
	)
	if err != nil {
		s.logger.Error("failed to update file status and ETag after multipart completion", zap.Error(err))
		return nil, errs.ErrInternal // Or a more specific error
	}

	file.Status = model.FileStatusCompleted
	file.ETag = &etag

	return &CompleteMultipartUploadResponse{
		File: file,
	}, nil
}

func (s *fileService) AbortMultipartUpload(
	ctx context.Context,
	req *AbortMultipartUploadRequest,
) (*AbortMultipartUploadResponse, error) {
	file, err := s.fileRepository.GetByObjectKey(ctx, req.ObjectKey) // Assuming GetByObjectKey exists or needs to be created
	if err != nil {
		return nil, err
	}

	if file.UploadID == nil || *file.UploadID != req.UploadID {
		return nil, errs.ErrMultipartUploadNotFound
	}

	_, err = s.storage.AbortMultipartUpload(
		ctx,
		storage.AbortMultipartUploadOptions{
			Bucket:    file.Bucket,
			ObjectKey: file.ObjectKey,
			UploadID:  req.UploadID,
		},
	)
	if err != nil {
		s.logger.Error("failed to abort multipart upload", zap.Error(err))
		return nil, errs.ErrAbortMultipartUpload
	}

	// Update file status to aborted/failed and clean up upload ID
	abortedStatus := model.FileStatusFailed // Or a new FileStatusAborted
	err = s.fileRepository.UpdateStatusAndClearUploadID(
		ctx,
		file.ID,
		abortedStatus,
	)
	if err != nil {
		s.logger.Error("failed to update file status after multipart abort", zap.Error(err))
		return nil, errs.ErrInternal
	}

	return &AbortMultipartUploadResponse{}, nil
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
