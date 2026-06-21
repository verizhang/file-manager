package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/verizhang/file-manager/internal/config"
	"github.com/verizhang/file-manager/internal/mocks"
	"github.com/verizhang/file-manager/internal/model"
	"github.com/verizhang/file-manager/internal/service"
	"github.com/verizhang/file-manager/pkg/errs"
	"github.com/verizhang/file-manager/pkg/storage"
	"github.com/verizhang/file-manager/pkg/storage/s3"
)

func TestFileService_CreateUploadUrl(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockFileRepo := mocks.NewMockFileRepository(ctrl)
	logger, _ := zap.NewDevelopment()

	cfg := &config.Config{
		S3: s3.Config{
			Bucket: "test-bucket",
		},
		File: config.FileConfig{
			AllowedTypes: []string{"image/jpeg", "image/png"},
			MaxFileSize:  10 * 1024 * 1024, // 10MB
		},
		PresignedConfig: config.PresignedConfig{
			UploadExpireMinutes: 5,
		},
	}

	fileService := service.NewFileService(cfg, logger, mockStorage, mockFileRepo)

	ctx := context.Background()

	t.Run("success - valid request", func(t *testing.T) {
		req := &service.CreateUploadRequest{
			FileName:    "test.jpg",
			ContentType: "image/jpeg",
			Size:        1024 * 1024, // 1MB
		}

		mockStorage.EXPECT().GeneratePresignedUploadURL(
			ctx,
			gomock.Any(), // We'll check the options more specifically below if needed
		).Return(&storage.GeneratePresignedUploadURLResult{URL: "http://presigned.url/test.jpg"}, nil)

		mockFileRepo.EXPECT().Create(
			ctx,
			gomock.Any(), // We'll check the file object more specifically below if needed
		).Return(nil)

		resp, err := fileService.CreateUploadUrl(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "http://presigned.url/test.jpg", resp.UploadURL)
		assert.NotNil(t, resp.File)
		assert.Equal(t, req.FileName, resp.File.FileName)
		assert.Equal(t, req.ContentType, resp.File.ContentType)
		assert.Equal(t, req.Size, resp.File.Size)
		assert.Equal(t, model.FileStatusPending, resp.File.Status)
		assert.Equal(t, model.VirusScanStatusPending, resp.File.VirusScanStatus)
		assert.NotEmpty(t, resp.File.ID)
		assert.NotEmpty(t, resp.File.ObjectKey)
	})

	t.Run("failure - invalid content type", func(t *testing.T) {
		req := &service.CreateUploadRequest{
			FileName:    "test.txt",
			ContentType: "text/plain",
			Size:        100,
		}

		resp, err := fileService.CreateUploadUrl(ctx, req)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrFileTypeNotAllowed))
		assert.Nil(t, resp)
	})

	t.Run("failure - file too large", func(t *testing.T) {
		req := &service.CreateUploadRequest{
			FileName:    "large.jpg",
			ContentType: "image/jpeg",
			Size:        20 * 1024 * 1024, // 20MB, exceeds 10MB limit
		}

		resp, err := fileService.CreateUploadUrl(ctx, req)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrFileTooLarge))
		assert.Nil(t, resp)
	})

	t.Run("failure - storage generates presigned URL error", func(t *testing.T) {
		req := &service.CreateUploadRequest{
			FileName:    "test.jpg",
			ContentType: "image/jpeg",
			Size:        1024,
		}
		expectedErr := errors.New("storage error")

		mockStorage.EXPECT().GeneratePresignedUploadURL(
			ctx,
			gomock.Any(),
		).Return(nil, expectedErr)

		resp, err := fileService.CreateUploadUrl(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})

	t.Run("failure - file repository create error", func(t *testing.T) {
		req := &service.CreateUploadRequest{
			FileName:    "test.jpg",
			ContentType: "image/jpeg",
			Size:        1024,
		}
		expectedErr := errors.New("repository error")

		mockStorage.EXPECT().GeneratePresignedUploadURL(
			ctx,
			gomock.Any(),
		).Return(&storage.GeneratePresignedUploadURLResult{URL: "http://presigned.url/test.jpg"}, nil)

		mockFileRepo.EXPECT().Create(
			ctx,
			gomock.Any(),
		).Return(expectedErr)

		resp, err := fileService.CreateUploadUrl(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})
}

func TestFileService_CompleteUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockFileRepo := mocks.NewMockFileRepository(ctrl)
	logger, _ := zap.NewDevelopment()

	cfg := &config.Config{
		S3: s3.Config{
			Bucket: "test-bucket",
		},
	}

	fileService := service.NewFileService(cfg, logger, mockStorage, mockFileRepo)

	ctx := context.Background()
	fileID := "test-file-id"
	objectKey := "2026/05/31/test-file-id.jpg"
	bucket := "test-bucket"

	t.Run("success - file pending, completes upload", func(t *testing.T) {
		req := &service.CompleteUploadRequest{FileID: fileID}
		file := &model.File{
			ID:        fileID,
			ObjectKey: objectKey,
			Bucket:    bucket,
			Status:    model.FileStatusPending,
		}

		mockFileRepo.EXPECT().GetByID(ctx, fileID).Return(file, nil)
		mockStorage.EXPECT().HeadObject(ctx, bucket, objectKey).Return(nil)
		mockFileRepo.EXPECT().UpdateStatus(ctx, fileID, model.FileStatusCompleted).Return(nil)

		resp, err := fileService.CompleteUpload(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, model.FileStatusCompleted, resp.File.Status)
	})

	t.Run("success - file already completed", func(t *testing.T) {
		req := &service.CompleteUploadRequest{FileID: fileID}
		file := &model.File{
			ID:        fileID,
			ObjectKey: objectKey,
			Bucket:    bucket,
			Status:    model.FileStatusCompleted,
		}

		mockFileRepo.EXPECT().GetByID(ctx, fileID).Return(file, nil)
		// No further calls to storage or fileRepo.UpdateStatus expected
		mockStorage.EXPECT().HeadObject(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		mockFileRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		resp, err := fileService.CompleteUpload(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, model.FileStatusCompleted, resp.File.Status)
	})

	t.Run("failure - file not found", func(t *testing.T) {
		req := &service.CompleteUploadRequest{FileID: fileID}
		expectedErr := errs.ErrFileNotFound

		mockFileRepo.EXPECT().GetByID(ctx, fileID).Return(nil, expectedErr)
		mockStorage.EXPECT().HeadObject(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		mockFileRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		resp, err := fileService.CompleteUpload(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})

	t.Run("failure - storage HeadObject fails", func(t *testing.T) {
		req := &service.CompleteUploadRequest{FileID: fileID}
		file := &model.File{
			ID:        fileID,
			ObjectKey: objectKey,
			Bucket:    bucket,
			Status:    model.FileStatusPending,
		}
		expectedErr := errors.New("head object error")

		mockFileRepo.EXPECT().GetByID(ctx, fileID).Return(file, nil)
		mockStorage.EXPECT().HeadObject(ctx, bucket, objectKey).Return(expectedErr)
		mockFileRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		resp, err := fileService.CompleteUpload(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})

	t.Run("failure - file repository UpdateStatus fails", func(t *testing.T) {
		req := &service.CompleteUploadRequest{FileID: fileID}
		file := &model.File{
			ID:        fileID,
			ObjectKey: objectKey,
			Bucket:    bucket,
			Status:    model.FileStatusPending,
		}
		expectedErr := errors.New("update status error")

		mockFileRepo.EXPECT().GetByID(ctx, fileID).Return(file, nil)
		mockStorage.EXPECT().HeadObject(ctx, bucket, objectKey).Return(nil)
		mockFileRepo.EXPECT().UpdateStatus(ctx, fileID, model.FileStatusCompleted).Return(expectedErr)

		resp, err := fileService.CompleteUpload(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})
}

func TestFileService_CreateMultipartUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockFileRepo := mocks.NewMockFileRepository(ctrl)
	logger, _ := zap.NewDevelopment()

	cfg := &config.Config{
		S3: s3.Config{
			Bucket: "test-bucket",
		},
		File: config.FileConfig{
			AllowedTypes: []string{"video/mp4"},
			MaxFileSize:  100 * 1024 * 1024, // 100MB
		},
		Multipart: config.MultipartConfig{
			PartSize: 5 * 1024 * 1024, // 5MB
		},
	}

	fileService := service.NewFileService(cfg, logger, mockStorage, mockFileRepo)

	ctx := context.Background()
	fileName := "test.mp4"
	contentType := "video/mp4"
	fileSize := int64(50 * 1024 * 1024) // 50MB

	t.Run("success - valid request", func(t *testing.T) {
		req := &service.CreateMultipartUploadRequest{
			FileName:    fileName,
			ContentType: contentType,
			Size:        fileSize,
		}
		uploadID := "test-upload-id"

		mockStorage.EXPECT().CreateMultipartUpload(
			ctx,
			gomock.Any(),
		).Return(&storage.CreateMultipartUploadResult{UploadID: uploadID}, nil)

		mockFileRepo.EXPECT().Create(
			ctx,
			gomock.Any(),
		).Return(nil)

		resp, err := fileService.CreateMultipartUpload(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.FileID)
		assert.Equal(t, uploadID, resp.UploadID)
		assert.NotEmpty(t, resp.ObjectKey)
		assert.Equal(t, cfg.Multipart.PartSize, resp.PartSize)
		assert.Equal(t, int32(10), resp.TotalParts) // 50MB / 5MB = 10 parts
	})

	t.Run("failure - invalid content type", func(t *testing.T) {
		req := &service.CreateMultipartUploadRequest{
			FileName:    "test.txt",
			ContentType: "text/plain",
			Size:        100,
		}

		resp, err := fileService.CreateMultipartUpload(ctx, req)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrFileTypeNotAllowed))
		assert.Nil(t, resp)
	})

	t.Run("failure - file too large", func(t *testing.T) {
		req := &service.CreateMultipartUploadRequest{
			FileName:    fileName,
			ContentType: contentType,
			Size:        200 * 1024 * 1024, // 200MB, exceeds 100MB limit
		}

		resp, err := fileService.CreateMultipartUpload(ctx, req)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrFileTooLarge))
		assert.Nil(t, resp)
	})

	t.Run("failure - storage CreateMultipartUpload fails", func(t *testing.T) {
		req := &service.CreateMultipartUploadRequest{
			FileName:    fileName,
			ContentType: contentType,
			Size:        fileSize,
		}
		expectedErr := errors.New("storage multipart error")

		mockStorage.EXPECT().CreateMultipartUpload(
			ctx,
			gomock.Any(),
		).Return(nil, expectedErr)
		mockFileRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)

		resp, err := fileService.CreateMultipartUpload(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})

	t.Run("failure - file repository Create fails, aborts multipart", func(t *testing.T) {
		req := &service.CreateMultipartUploadRequest{
			FileName:    fileName,
			ContentType: contentType,
			Size:        fileSize,
		}
		uploadID := "test-upload-id"
		expectedErr := errors.New("repository create error")

		mockStorage.EXPECT().CreateMultipartUpload(
			ctx,
			gomock.Any(),
		).Return(&storage.CreateMultipartUploadResult{UploadID: uploadID}, nil)

		mockFileRepo.EXPECT().Create(
			ctx,
			gomock.Any(),
		).Return(expectedErr)

		mockStorage.EXPECT().AbortMultipartUpload(
			ctx,
			gomock.Any(),
		).Return(&storage.AbortMultipartUploadResult{}, nil) // Expect abort to be called

		resp, err := fileService.CreateMultipartUpload(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})

	t.Run("failure - file repository Create fails, aborts multipart also fails", func(t *testing.T) {
		req := &service.CreateMultipartUploadRequest{
			FileName:    fileName,
			ContentType: contentType,
			Size:        fileSize,
		}
		uploadID := "test-upload-id"
		expectedCreateErr := errors.New("repository create error")
		expectedAbortErr := errors.New("abort multipart error")

		mockStorage.EXPECT().CreateMultipartUpload(
			ctx,
			gomock.Any(),
		).Return(&storage.CreateMultipartUploadResult{UploadID: uploadID}, nil)

		mockFileRepo.EXPECT().Create(
			ctx,
			gomock.Any(),
		).Return(expectedCreateErr)

		mockStorage.EXPECT().AbortMultipartUpload(
			ctx,
			gomock.Any(),
		).Return(nil, expectedAbortErr) // Expect abort to be called and fail

		resp, err := fileService.CreateMultipartUpload(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedCreateErr, err) // Original error should be returned
		assert.Nil(t, resp)
	})
}

func TestFileService_CreateMultipartUploadUrl(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockFileRepo := mocks.NewMockFileRepository(ctrl)
	logger, _ := zap.NewDevelopment()

	cfg := &config.Config{
		S3: s3.Config{
			Bucket: "test-bucket",
		},
		PresignedConfig: config.PresignedConfig{
			UploadExpireMinutes: 5,
		},
	}

	fileService := service.NewFileService(cfg, logger, mockStorage, mockFileRepo)

	ctx := context.Background()
	fileID := "test-file-id"
	uploadID := "test-upload-id"
	objectKey := "2026/05/31/test-file-id.mp4"
	contentType := "video/mp4"
	partNumber := int32(1)

	t.Run("success - valid request, file pending", func(t *testing.T) {
		req := &service.CreateMultipartUploadUrlRequest{
			FileID:     fileID,
			UploadID:   uploadID,
			ObjectKey:  objectKey,
			PartNumber: partNumber,
		}
		file := &model.File{
			ID:          fileID,
			UploadID:    &uploadID,
			ObjectKey:   objectKey,
			ContentType: contentType,
			Status:      model.FileStatusPending,
		}
		presignedURL := "http://presigned.url/part1"

		mockFileRepo.EXPECT().GetByID(ctx, fileID).Return(file, nil)
		mockStorage.EXPECT().GeneratePresignedMultipartUploadURL(
			ctx,
			gomock.Any(),
		).Return(&storage.GeneratePresignedMultipartUploadURLResult{URL: presignedURL, Headers: map[string]string{}}, nil)
		mockFileRepo.EXPECT().UpdateStatus(ctx, fileID, model.FileStatusCompleted).Return(nil) // Expect status update

		resp, err := fileService.CreateMultipartUploadUrl(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, presignedURL, resp.UploadURL)
	})

	t.Run("success - valid request, file uploading", func(t *testing.T) {
		req := &service.CreateMultipartUploadUrlRequest{
			FileID:     fileID,
			UploadID:   uploadID,
			ObjectKey:  objectKey,
			PartNumber: partNumber,
		}
		file := &model.File{
			ID:          fileID,
			UploadID:    &uploadID,
			ObjectKey:   objectKey,
			ContentType: contentType,
			Status:      model.FileStatusUploading, // Status is already uploading
		}
		presignedURL := "http://presigned.url/part1"

		mockFileRepo.EXPECT().GetByID(ctx, fileID).Return(file, nil)
		mockStorage.EXPECT().GeneratePresignedMultipartUploadURL(
			ctx,
			gomock.Any(),
		).Return(&storage.GeneratePresignedMultipartUploadURLResult{URL: presignedURL, Headers: map[string]string{}}, nil)
		mockFileRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0) // No status update expected

		resp, err := fileService.CreateMultipartUploadUrl(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, presignedURL, resp.UploadURL)
	})

	t.Run("failure - file not found", func(t *testing.T) {
		req := &service.CreateMultipartUploadUrlRequest{
			FileID:     fileID,
			UploadID:   uploadID,
			ObjectKey:  objectKey,
			PartNumber: partNumber,
		}
		expectedErr := errs.ErrFileNotFound

		mockFileRepo.EXPECT().GetByID(ctx, fileID).Return(nil, expectedErr)
		mockStorage.EXPECT().GeneratePresignedMultipartUploadURL(gomock.Any(), gomock.Any()).Times(0)
		mockFileRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		resp, err := fileService.CreateMultipartUploadUrl(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})

	t.Run("failure - upload ID mismatch", func(t *testing.T) {
		req := &service.CreateMultipartUploadUrlRequest{
			FileID:     fileID,
			UploadID:   "wrong-upload-id",
			ObjectKey:  objectKey,
			PartNumber: partNumber,
		}
		file := &model.File{
			ID:        fileID,
			UploadID:  &uploadID,
			ObjectKey: objectKey,
		}

		mockFileRepo.EXPECT().GetByID(ctx, fileID).Return(file, nil)
		mockStorage.EXPECT().GeneratePresignedMultipartUploadURL(gomock.Any(), gomock.Any()).Times(0)
		mockFileRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		resp, err := fileService.CreateMultipartUploadUrl(ctx, req)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrMultipartUploadMismatch))
		assert.Nil(t, resp)
	})

	t.Run("failure - object key mismatch", func(t *testing.T) {
		req := &service.CreateMultipartUploadUrlRequest{
			FileID:     fileID,
			UploadID:   uploadID,
			ObjectKey:  "wrong-object-key",
			PartNumber: partNumber,
		}
		file := &model.File{
			ID:        fileID,
			UploadID:  &uploadID,
			ObjectKey: objectKey,
		}

		mockFileRepo.EXPECT().GetByID(ctx, fileID).Return(file, nil)
		mockStorage.EXPECT().GeneratePresignedMultipartUploadURL(gomock.Any(), gomock.Any()).Times(0)
		mockFileRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		resp, err := fileService.CreateMultipartUploadUrl(ctx, req)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrMultipartUploadMismatch))
		assert.Nil(t, resp)
	})

	t.Run("failure - storage GeneratePresignedMultipartUploadURL fails", func(t *testing.T) {
		req := &service.CreateMultipartUploadUrlRequest{
			FileID:     fileID,
			UploadID:   uploadID,
			ObjectKey:  objectKey,
			PartNumber: partNumber,
		}
		file := &model.File{
			ID:          fileID,
			UploadID:    &uploadID,
			ObjectKey:   objectKey,
			ContentType: contentType,
			Status:      model.FileStatusPending,
		}
		expectedErr := errors.New("presigned url error")

		mockFileRepo.EXPECT().GetByID(ctx, fileID).Return(file, nil)
		mockStorage.EXPECT().GeneratePresignedMultipartUploadURL(
			ctx,
			gomock.Any(),
		).Return(nil, expectedErr)
		mockFileRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0) // Status update should not happen if presigned URL fails

		resp, err := fileService.CreateMultipartUploadUrl(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})

	t.Run("failure - file repository UpdateStatus fails after presigned URL generation", func(t *testing.T) {
		req := &service.CreateMultipartUploadUrlRequest{
			FileID:     fileID,
			UploadID:   uploadID,
			ObjectKey:  objectKey,
			PartNumber: partNumber,
		}
		file := &model.File{
			ID:          fileID,
			UploadID:    &uploadID,
			ObjectKey:   objectKey,
			ContentType: contentType,
			Status:      model.FileStatusPending,
		}
		presignedURL := "http://presigned.url/part1"
		expectedErr := errors.New("update status error")

		mockFileRepo.EXPECT().GetByID(ctx, fileID).Return(file, nil)
		mockStorage.EXPECT().GeneratePresignedMultipartUploadURL(
			ctx,
			gomock.Any(),
		).Return(&storage.GeneratePresignedMultipartUploadURLResult{URL: presignedURL, Headers: map[string]string{}}, nil)
		mockFileRepo.EXPECT().UpdateStatus(ctx, fileID, model.FileStatusCompleted).Return(expectedErr)

		resp, err := fileService.CreateMultipartUploadUrl(ctx, req)

		assert.NoError(t, err) // The error from UpdateStatus is logged but not returned
		assert.NotNil(t, resp)
		assert.Equal(t, presignedURL, resp.UploadURL)
	})
}

func TestFileService_CompleteMultipartUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockFileRepo := mocks.NewMockFileRepository(ctrl)
	logger, _ := zap.NewDevelopment()

	cfg := &config.Config{
		S3: s3.Config{
			Bucket: "test-bucket",
		},
	}

	fileService := service.NewFileService(cfg, logger, mockStorage, mockFileRepo)

	ctx := context.Background()
	objectKey := "2026/05/31/test-file-id.mp4"
	uploadID := "test-upload-id"
	fileID := "test-file-id"
	etag := "test-etag"

	parts := []service.MultipartPart{
		{PartNumber: 1, ETag: "etag1"},
		{PartNumber: 2, ETag: "etag2"},
	}

	t.Run("success - valid request, file pending", func(t *testing.T) {
		req := &service.CompleteMultipartUploadRequest{
			UploadID:  uploadID,
			ObjectKey: objectKey,
			Parts:     parts,
		}
		file := &model.File{
			ID:        fileID,
			UploadID:  &uploadID,
			ObjectKey: objectKey,
			Status:    model.FileStatusPending,
		}

		mockFileRepo.EXPECT().GetByObjectKey(ctx, objectKey).Return(file, nil)
		mockStorage.EXPECT().CompleteMultipartUpload(
			ctx,
			gomock.Any(),
		).Return(&storage.CompleteMultipartUploadResult{ETag: etag}, nil)
		mockFileRepo.EXPECT().UpdateStatusAndETag(ctx, fileID, model.FileStatusCompleted, etag).Return(nil)

		resp, err := fileService.CompleteMultipartUpload(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, model.FileStatusCompleted, resp.File.Status)
		assert.Equal(t, etag, *resp.File.ETag)
	})

	t.Run("success - file already completed", func(t *testing.T) {
		req := &service.CompleteMultipartUploadRequest{
			UploadID:  uploadID,
			ObjectKey: objectKey,
			Parts:     parts,
		}
		file := &model.File{
			ID:        fileID,
			UploadID:  &uploadID,
			ObjectKey: objectKey,
			Status:    model.FileStatusCompleted,
			ETag:      &etag,
		}

		mockFileRepo.EXPECT().GetByObjectKey(ctx, objectKey).Return(file, nil)
		mockStorage.EXPECT().CompleteMultipartUpload(gomock.Any(), gomock.Any()).Times(0)
		mockFileRepo.EXPECT().UpdateStatusAndETag(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		resp, err := fileService.CompleteMultipartUpload(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, model.FileStatusCompleted, resp.File.Status)
		assert.Equal(t, etag, *resp.File.ETag)
	})

	t.Run("failure - file not found", func(t *testing.T) {
		req := &service.CompleteMultipartUploadRequest{
			UploadID:  uploadID,
			ObjectKey: objectKey,
			Parts:     parts,
		}
		expectedErr := errs.ErrFileNotFound

		mockFileRepo.EXPECT().GetByObjectKey(ctx, objectKey).Return(nil, expectedErr)
		mockStorage.EXPECT().CompleteMultipartUpload(gomock.Any(), gomock.Any()).Times(0)
		mockFileRepo.EXPECT().UpdateStatusAndETag(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		resp, err := fileService.CompleteMultipartUpload(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})

	t.Run("failure - upload ID mismatch", func(t *testing.T) {
		req := &service.CompleteMultipartUploadRequest{
			UploadID:  "wrong-upload-id",
			ObjectKey: objectKey,
			Parts:     parts,
		}
		file := &model.File{
			ID:        fileID,
			UploadID:  &uploadID,
			ObjectKey: objectKey,
			Status:    model.FileStatusPending,
		}

		mockFileRepo.EXPECT().GetByObjectKey(ctx, objectKey).Return(file, nil)
		mockStorage.EXPECT().CompleteMultipartUpload(gomock.Any(), gomock.Any()).Times(0)
		mockFileRepo.EXPECT().UpdateStatusAndETag(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		resp, err := fileService.CompleteMultipartUpload(ctx, req)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrMultipartUploadMismatch))
		assert.Nil(t, resp)
	})

	t.Run("failure - storage CompleteMultipartUpload fails", func(t *testing.T) {
		req := &service.CompleteMultipartUploadRequest{
			UploadID:  uploadID,
			ObjectKey: objectKey,
			Parts:     parts,
		}
		file := &model.File{
			ID:        fileID,
			UploadID:  &uploadID,
			ObjectKey: objectKey,
			Status:    model.FileStatusPending,
		}
		expectedErr := errors.New("storage complete multipart error")

		mockFileRepo.EXPECT().GetByObjectKey(ctx, objectKey).Return(file, nil)
		mockStorage.EXPECT().CompleteMultipartUpload(
			ctx,
			gomock.Any(),
		).Return(nil, expectedErr)
		mockFileRepo.EXPECT().UpdateStatusAndETag(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		resp, err := fileService.CompleteMultipartUpload(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})

	t.Run("failure - file repository UpdateStatusAndETag fails", func(t *testing.T) {
		req := &service.CompleteMultipartUploadRequest{
			UploadID:  uploadID,
			ObjectKey: objectKey,
			Parts:     parts,
		}
		file := &model.File{
			ID:        fileID,
			UploadID:  &uploadID,
			ObjectKey: objectKey,
			Status:    model.FileStatusPending,
		}
		expectedErr := errors.New("update status and etag error")

		mockFileRepo.EXPECT().GetByObjectKey(ctx, objectKey).Return(file, nil)
		mockStorage.EXPECT().CompleteMultipartUpload(
			ctx,
			gomock.Any(),
		).Return(&storage.CompleteMultipartUploadResult{ETag: etag}, nil)
		mockFileRepo.EXPECT().UpdateStatusAndETag(ctx, fileID, model.FileStatusCompleted, etag).Return(expectedErr)

		resp, err := fileService.CompleteMultipartUpload(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})
}

func TestFileService_AbortMultipartUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockFileRepo := mocks.NewMockFileRepository(ctrl)
	logger, _ := zap.NewDevelopment()

	cfg := &config.Config{
		S3: s3.Config{
			Bucket: "test-bucket",
		},
	}

	fileService := service.NewFileService(cfg, logger, mockStorage, mockFileRepo)

	ctx := context.Background()
	objectKey := "2026/05/31/test-file-id.mp4"
	uploadID := "test-upload-id"
	fileID := "test-file-id"

	t.Run("success - valid request", func(t *testing.T) {
		req := &service.AbortMultipartUploadRequest{
			UploadID:  uploadID,
			ObjectKey: objectKey,
		}
		file := &model.File{
			ID:        fileID,
			UploadID:  &uploadID,
			ObjectKey: objectKey,
			Status:    model.FileStatusUploading,
		}

		mockFileRepo.EXPECT().GetByObjectKey(ctx, objectKey).Return(file, nil)
		mockStorage.EXPECT().AbortMultipartUpload(
			ctx,
			gomock.Any(),
		).Return(&storage.AbortMultipartUploadResult{}, nil)
		mockFileRepo.EXPECT().UpdateStatusAndClearUploadID(ctx, fileID, model.FileStatusAborted).Return(nil)

		resp, err := fileService.AbortMultipartUpload(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("failure - file not found", func(t *testing.T) {
		req := &service.AbortMultipartUploadRequest{
			UploadID:  uploadID,
			ObjectKey: objectKey,
		}
		expectedErr := errs.ErrFileNotFound

		mockFileRepo.EXPECT().GetByObjectKey(ctx, objectKey).Return(nil, expectedErr)
		mockStorage.EXPECT().AbortMultipartUpload(gomock.Any(), gomock.Any()).Times(0)
		mockFileRepo.EXPECT().UpdateStatusAndClearUploadID(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		resp, err := fileService.AbortMultipartUpload(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})

	t.Run("failure - upload ID mismatch", func(t *testing.T) {
		req := &service.AbortMultipartUploadRequest{
			UploadID:  "wrong-upload-id",
			ObjectKey: objectKey,
		}
		file := &model.File{
			ID:        fileID,
			UploadID:  &uploadID,
			ObjectKey: objectKey,
			Status:    model.FileStatusUploading,
		}

		mockFileRepo.EXPECT().GetByObjectKey(ctx, objectKey).Return(file, nil)
		mockStorage.EXPECT().AbortMultipartUpload(gomock.Any(), gomock.Any()).Times(0)
		mockFileRepo.EXPECT().UpdateStatusAndClearUploadID(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		resp, err := fileService.AbortMultipartUpload(ctx, req)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrMultipartUploadMismatch))
		assert.Nil(t, resp)
	})

	t.Run("failure - storage AbortMultipartUpload fails", func(t *testing.T) {
		req := &service.AbortMultipartUploadRequest{
			UploadID:  uploadID,
			ObjectKey: objectKey,
		}
		file := &model.File{
			ID:        fileID,
			UploadID:  &uploadID,
			ObjectKey: objectKey,
			Status:    model.FileStatusUploading,
		}
		expectedErr := errors.New("storage abort error")

		mockFileRepo.EXPECT().GetByObjectKey(ctx, objectKey).Return(file, nil)
		mockStorage.EXPECT().AbortMultipartUpload(
			ctx,
			gomock.Any(),
		).Return(nil, expectedErr)
		mockFileRepo.EXPECT().UpdateStatusAndClearUploadID(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		resp, err := fileService.AbortMultipartUpload(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})

	t.Run("failure - file repository UpdateStatusAndClearUploadID fails", func(t *testing.T) {
		req := &service.AbortMultipartUploadRequest{
			UploadID:  uploadID,
			ObjectKey: objectKey,
		}
		file := &model.File{
			ID:        fileID,
			UploadID:  &uploadID,
			ObjectKey: objectKey,
			Status:    model.FileStatusUploading,
		}
		expectedErr := errors.New("update status and clear upload id error")

		mockFileRepo.EXPECT().GetByObjectKey(ctx, objectKey).Return(file, nil)
		mockStorage.EXPECT().AbortMultipartUpload(
			ctx,
			gomock.Any(),
		).Return(&storage.AbortMultipartUploadResult{}, nil)
		mockFileRepo.EXPECT().UpdateStatusAndClearUploadID(ctx, fileID, model.FileStatusAborted).Return(expectedErr)

		resp, err := fileService.AbortMultipartUpload(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})
}

func TestFileService_GetFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockFileRepo := mocks.NewMockFileRepository(ctrl)
	logger, _ := zap.NewDevelopment()

	cfg := &config.Config{} // No specific config needed for GetFile

	fileService := service.NewFileService(cfg, logger, mockStorage, mockFileRepo)

	ctx := context.Background()
	fileID := "test-file-id"

	t.Run("success - file found", func(t *testing.T) {
		req := &service.GetFileRequest{FileID: fileID}
		expectedFile := &model.File{
			ID:          fileID,
			FileName:    "test.jpg",
			ContentType: "image/jpeg",
			Size:        1024,
			Status:      model.FileStatusCompleted,
		}

		mockFileRepo.EXPECT().GetByID(ctx, fileID).Return(expectedFile, nil)

		resp, err := fileService.GetFile(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedFile, resp.File)
	})

	t.Run("failure - file not found", func(t *testing.T) {
		req := &service.GetFileRequest{FileID: fileID}
		expectedErr := errs.ErrFileNotFound

		mockFileRepo.EXPECT().GetByID(ctx, fileID).Return(nil, expectedErr)

		resp, err := fileService.GetFile(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})
}
