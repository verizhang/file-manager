package service

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/verizhang/file-manager/internal/config"
	"github.com/verizhang/file-manager/internal/errs"
	"github.com/verizhang/file-manager/internal/messaging"
	"github.com/verizhang/file-manager/internal/model"
	"github.com/verizhang/file-manager/internal/repository"
	"github.com/verizhang/file-manager/internal/storage"
	virusscanner "github.com/verizhang/file-manager/internal/virus-scanner"
)

const (
	COMPLETE_UPLOAD_TOPIC = "complete_upload"
)

type fileService struct {
	cfg            *config.Config
	logger         *zap.Logger
	storage        storage.Storage
	fileRepository repository.FileRepository
	messaging      messaging.Messaging
	virusScanner   virusscanner.Scanner
}

func NewFileService(
	cfg *config.Config,
	logger *zap.Logger,
	storage storage.Storage,
	fileRepository repository.FileRepository,
	deps ...interface{},
) FileService {
	var messagingClient messaging.Messaging
	var virusScanner virusscanner.Scanner

	for _, dep := range deps {
		switch d := dep.(type) {
		case messaging.Messaging:
			messagingClient = d
		case virusscanner.Scanner:
			virusScanner = d
		}
	}

	return &fileService{
		cfg:            cfg,
		logger:         logger,
		storage:        storage,
		fileRepository: fileRepository,
		messaging:      messagingClient,
		virusScanner:   virusScanner,
	}
}

func (s *fileService) CreateUploadUrl(
	ctx context.Context,
	req *CreateUploadRequest,
) (*CreateUploadResponse, error) {

	err := s.validateContentType(req.ContentType)
	if err != nil {
		return nil, err
	}

	err = s.validateFileSize(req.Size)
	if err != nil {
		return nil, err
	}

	fileID := uuid.NewString()

	objectKey := GenerateObjectKey(
		req.UserID,
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
		return nil, err
	}

	file := &model.File{
		ID:              fileID,
		ObjectKey:       objectKey,
		Bucket:          s.cfg.S3.Bucket,
		FileName:        req.FileName,
		ContentType:     req.ContentType,
		Size:            req.Size,
		Status:          model.FileStatusPending,
		VirusScanStatus: model.VirusScanStatusPending,
	}

	err = s.fileRepository.Create(
		ctx,
		file,
	)
	if err != nil {
		return nil, err
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

	// Publish complete upload file
	message, err := json.Marshal(file)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrInternal, err)
	}

	if s.messaging != nil {
		err = s.messaging.Publish(ctx, COMPLETE_UPLOAD_TOPIC, message)
		if err != nil {
			s.logger.Error("failed to publish complete upload file", zap.Error(err))
		}
	}

	return &CompleteUploadResponse{
		File: file,
	}, nil
}

func (s *fileService) CreateMultipartUpload(
	ctx context.Context,
	req *CreateMultipartUploadRequest,
) (*CreateMultipartUploadResponse, error) {
	err := s.validateContentType(req.ContentType)
	if err != nil {
		return nil, err
	}

	err = s.validateFileSize(req.Size)
	if err != nil {
		return nil, err
	}

	fileID := uuid.NewString()

	objectKey := GenerateObjectKey(
		req.UserID,
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
		return nil, err
	}

	file := &model.File{
		ID:              fileID,
		UploadID:        &createMultipartOutput.UploadID,
		ObjectKey:       objectKey,
		Bucket:          s.cfg.S3.Bucket,
		FileName:        req.FileName,
		ContentType:     req.ContentType,
		Size:            req.Size,
		Status:          model.FileStatusPending,
		VirusScanStatus: model.VirusScanStatusPending,
	}

	err = s.fileRepository.Create(
		ctx,
		file,
	)
	if err != nil {
		// Attempt to abort the multipart upload if metadata creation fails
		_, abortErr := s.storage.AbortMultipartUpload(ctx, storage.AbortMultipartUploadOptions{
			Bucket:    s.cfg.S3.Bucket,
			ObjectKey: objectKey,
			UploadID:  createMultipartOutput.UploadID,
		})
		if abortErr != nil {
			s.logger.Error("failed to abort multipart upload after metadata creation failed", zap.Error(abortErr))
		}
		return nil, err
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
		return nil, fmt.Errorf("%w: %s", errs.ErrMultipartUploadMismatch, "upload id missmatch")
	}

	if file.ObjectKey != req.ObjectKey {
		return nil, fmt.Errorf("%w: %s", errs.ErrMultipartUploadMismatch, "object key missmatch")
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
		return nil, err
	}

	if file.Status != model.FileStatusUploading {
		err := s.fileRepository.UpdateStatus(
			ctx,
			file.ID,
			model.FileStatusCompleted,
		)
		if err != nil {
			s.logger.Error("failed to update status", zap.Error(err))
		}
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
		return nil, fmt.Errorf("%w: %s", errs.ErrMultipartUploadMismatch, "upload id missmatch")
	}

	if file.Status == model.FileStatusCompleted {
		return &CompleteMultipartUploadResponse{
			File: file,
		}, nil
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
		return nil, err
	}

	// Update file status and ETag
	etag := completeMultipartOutput.ETag
	err = s.fileRepository.UpdateStatusAndETag(
		ctx,
		file.ID,
		model.FileStatusCompleted,
		etag,
	)
	if err != nil {
		return nil, err
	}

	file.Status = model.FileStatusCompleted
	file.ETag = &etag

	// Publish complete upload file
	message, err := json.Marshal(file)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrInternal, err)
	}

	if s.messaging != nil {
		err = s.messaging.Publish(ctx, COMPLETE_UPLOAD_TOPIC, message)
		if err != nil {
			s.logger.Error("failed to publish complete upload file", zap.Error(err))
		}
	}

	return &CompleteMultipartUploadResponse{
		File: file,
	}, nil
}

func (s *fileService) AbortMultipartUpload(
	ctx context.Context,
	req *AbortMultipartUploadRequest,
) (*AbortMultipartUploadResponse, error) {
	file, err := s.fileRepository.GetByObjectKey(ctx, req.ObjectKey)
	if err != nil {
		return nil, err
	}

	if file.UploadID == nil || *file.UploadID != req.UploadID {
		return nil, fmt.Errorf("%w: %s", errs.ErrMultipartUploadMismatch, "upload id missmatch")
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
		return nil, err
	}

	abortedStatus := model.FileStatusAborted
	err = s.fileRepository.UpdateStatusAndClearUploadID(
		ctx,
		file.ID,
		abortedStatus,
	)
	if err != nil {
		return nil, err
	}

	return &AbortMultipartUploadResponse{}, nil
}

func (s *fileService) GetFile(
	ctx context.Context,
	req *GetFileRequest,
) (*GetFileResponse, error) {
	file, err := s.fileRepository.GetByID(
		ctx,
		req.FileID,
	)
	if err != nil {
		return nil, err
	}

	return &GetFileResponse{
		File: file,
	}, nil
}

func (s *fileService) CreateDownloadURL(
	ctx context.Context,
	req *CreateDownloadURLRequest,
) (*CreateDownloadURLResponse, error) {
	file, err := s.fileRepository.GetByID(
		ctx,
		req.FileID,
	)
	if err != nil {
		return nil, err
	}

	downloadURL, err := s.storage.GeneratePresignedDownloadURL(
		ctx,
		storage.GeneratePresignedDownloadURLOptions{
			Bucket:    file.Bucket,
			ObjectKey: file.ObjectKey,
			Expiry:    s.cfg.S3.PresignedConfig.DownloadExpireMinutes,
		},
	)
	if err != nil {
		return nil, err
	}

	return &CreateDownloadURLResponse{
		DownloadURL: downloadURL.URL,
	}, nil
}

func (s *fileService) DeleteFile(
	ctx context.Context,
	req *DeleteFileRequest,
) (*DeleteFileResponse, error) {
	file, err := s.fileRepository.GetByID(
		ctx,
		req.FileID,
	)
	if err != nil {
		return nil, err
	}

	_, err = s.storage.DeleteObject(
		ctx,
		file.Bucket,
		file.ObjectKey,
	)
	if err != nil {
		return nil, err
	}

	err = s.fileRepository.Delete(
		ctx,
		file.ID,
	)
	if err != nil {
		return nil, err
	}

	return &DeleteFileResponse{}, nil
}

func (s *fileService) ScanFile(
	ctx context.Context,
	file model.File,
) error {
	if err := s.fileRepository.UpdateVirusScanStatus(
		ctx,
		file.ID,
		model.VirusScanStatusScaning,
	); err != nil {
		return err
	}

	object, err := s.storage.GetObject(ctx, storage.GetObjectOptions{
		Bucket:    file.Bucket,
		ObjectKey: file.ObjectKey,
	})
	if err != nil {
		_ = s.fileRepository.UpdateVirusScanStatus(ctx, file.ID, model.VirusScanStatusFailed)
		return err
	}
	defer object.Body.Close()

	result, err := s.virusScanner.Scan(ctx, virusscanner.ScanOptions{
		FileName: file.FileName,
		Reader:   object.Body,
	})
	if err != nil {
		_ = s.fileRepository.UpdateVirusScanStatus(ctx, file.ID, model.VirusScanStatusFailed)
		return err
	}

	status := model.VirusScanStatusClean
	if result.Status == virusscanner.StatusInfected {
		status = model.VirusScanStatusInfected
	}

	if err := s.fileRepository.UpdateVirusScanStatus(
		ctx,
		file.ID,
		status,
	); err != nil {
		return err
	}

	s.logger.Info(
		"file virus scan completed",
		zap.String("file_id", file.ID),
		zap.String("status", string(status)),
		zap.String("threat", result.Threat),
	)

	return nil
}

func GenerateObjectKey(
	userID string,
	fileID string,
	fileName string,
) string {
	ext := filepath.Ext(fileName)

	return fmt.Sprintf(
		"%s/%s%s",
		userID,
		fileID,
		ext,
	)
}
