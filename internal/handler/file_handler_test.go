package grpc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	filev1 "github.com/verizhang/file-manager/gen/go/file/v1"
	handler "github.com/verizhang/file-manager/internal/handler" // Alias to avoid conflict with package name
	"github.com/verizhang/file-manager/internal/errs"
	"github.com/verizhang/file-manager/internal/mocks"
	"github.com/verizhang/file-manager/internal/model"
	"github.com/verizhang/file-manager/internal/service"
)

func TestFileHandler_CreateUploadUrl(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileService := mocks.NewMockFileService(ctrl)
	logger := zap.NewNop() // Use a no-op logger for tests

	fileHandler := handler.NewFileHandler(logger, mockFileService)

	ctx := context.Background()

	t.Run("success - valid request", func(t *testing.T) {
		req := &filev1.CreateUploadUrlRequest{
			FileName:    "test.jpg",
			ContentType: "image/jpeg",
			Size:        1024 * 1024,
		}

		expectedServiceResponse := &service.CreateUploadResponse{
			File: &model.File{
				ID:          "some-uuid",
				FileName:    req.FileName,
				ContentType: req.ContentType,
				Size:        req.Size,
			},
			UploadURL: "http://presigned.url/test.jpg",
		}

		mockFileService.EXPECT().CreateUploadUrl(
			ctx,
			&service.CreateUploadRequest{
				FileName:    req.FileName,
				ContentType: req.ContentType,
				Size:        req.Size,
			},
		).Return(expectedServiceResponse, nil)

		resp, err := fileHandler.CreateUploadUrl(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedServiceResponse.File.ID, resp.FileId)
		assert.Equal(t, expectedServiceResponse.UploadURL, resp.UploadUrl)
	})

	t.Run("failure - validation error (empty file name)", func(t *testing.T) {
		req := &filev1.CreateUploadUrlRequest{
			FileName:    "", // Invalid
			ContentType: "image/jpeg",
			Size:        1024,
		}

		resp, err := fileHandler.CreateUploadUrl(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Equal(t, "validation failed", st.Message()) // Corrected assertion
		assert.Nil(t, resp)
		mockFileService.EXPECT().CreateUploadUrl(gomock.Any(), gomock.Any()).Times(0) // Ensure service is not called
	})

	t.Run("failure - service returns error", func(t *testing.T) {
		req := &filev1.CreateUploadUrlRequest{
			FileName:    "test.jpg",
			ContentType: "image/jpeg",
			Size:        1024,
		}
		serviceErr := errors.New("something went wrong in service")

		mockFileService.EXPECT().CreateUploadUrl(
			ctx,
			&service.CreateUploadRequest{
				FileName:    req.FileName,
				ContentType: req.ContentType,
				Size:        req.Size,
			},
		).Return(nil, serviceErr)

		resp, err := fileHandler.CreateUploadUrl(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Equal(t, "internal server error", st.Message()) // Corrected assertion
		assert.Nil(t, resp)
	})

	t.Run("failure - service returns specific errs.ErrFileTypeNotAllowed", func(t *testing.T) {
		req := &filev1.CreateUploadUrlRequest{
			FileName:    "test.txt",
			ContentType: "text/plain",
			Size:        1024,
		}

		mockFileService.EXPECT().CreateUploadUrl(
			ctx,
			&service.CreateUploadRequest{
				FileName:    req.FileName,
				ContentType: req.ContentType,
				Size:        req.Size,
			},
		).Return(nil, errs.ErrFileTypeNotAllowed)

		resp, err := fileHandler.CreateUploadUrl(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), errs.ErrFileTypeNotAllowed.Error()) // Corrected assertion
		assert.Nil(t, resp)
	})
}

func TestFileHandler_CompleteUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileService := mocks.NewMockFileService(ctrl)
	logger := zap.NewNop()

	fileHandler := handler.NewFileHandler(logger, mockFileService)

	ctx := context.Background()
	fileID := "test-file-id"

	t.Run("success - valid request", func(t *testing.T) {
		req := &filev1.CompleteUploadRequest{FileId: fileID}
		expectedServiceResponse := &service.CompleteUploadResponse{
			File: &model.File{
				ID:     fileID,
				Status: model.FileStatusCompleted,
			},
		}

		mockFileService.EXPECT().CompleteUpload(
			ctx,
			&service.CompleteUploadRequest{FileID: fileID},
		).Return(expectedServiceResponse, nil)

		resp, err := fileHandler.CompleteUpload(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, fileID, resp.File.Id)
		assert.Equal(t, filev1.FileStatus_FILE_STATUS_COMPLETED, resp.File.Status)
	})

	t.Run("failure - validation error (empty file id)", func(t *testing.T) {
		req := &filev1.CompleteUploadRequest{FileId: ""} // Invalid

		resp, err := fileHandler.CompleteUpload(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Equal(t, "validation failed", st.Message())
		assert.Nil(t, resp)
		mockFileService.EXPECT().CompleteUpload(gomock.Any(), gomock.Any()).Times(0)
	})

	t.Run("failure - service returns error", func(t *testing.T) {
		req := &filev1.CompleteUploadRequest{FileId: fileID}
		serviceErr := errors.New("service complete upload error")

		mockFileService.EXPECT().CompleteUpload(
			ctx,
			&service.CompleteUploadRequest{FileID: fileID},
		).Return(nil, serviceErr)

		resp, err := fileHandler.CompleteUpload(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Equal(t, "internal server error", st.Message())
		assert.Nil(t, resp)
	})

	t.Run("failure - service returns specific errs.ErrFileNotFound", func(t *testing.T) {
		req := &filev1.CompleteUploadRequest{FileId: fileID}

		mockFileService.EXPECT().CompleteUpload(
			ctx,
			&service.CompleteUploadRequest{FileID: fileID},
		).Return(nil, errs.ErrFileNotFound)

		resp, err := fileHandler.CompleteUpload(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Contains(t, st.Message(), errs.ErrFileNotFound.Error())
		assert.Nil(t, resp)
	})
}

func TestFileHandler_CreateMultipartUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileService := mocks.NewMockFileService(ctrl)
	logger := zap.NewNop()

	fileHandler := handler.NewFileHandler(logger, mockFileService)

	ctx := context.Background()
	fileName := "test.mp4"
	contentType := "video/mp4"
	fileSize := int64(50 * 1024 * 1024)

	t.Run("success - valid request", func(t *testing.T) {
		req := &filev1.CreateMultipartUploadRequest{
			FileName:    fileName,
			ContentType: contentType,
			Size:        fileSize,
		}
		expectedServiceResponse := &service.CreateMultipartUploadResponse{
			FileID:     "some-file-id",
			UploadID:   "some-upload-id",
			ObjectKey:  "some-object-key",
			PartSize:   5 * 1024 * 1024,
			TotalParts: 10,
		}

		mockFileService.EXPECT().CreateMultipartUpload(
			ctx,
			&service.CreateMultipartUploadRequest{
				FileName:    fileName,
				ContentType: contentType,
				Size:        fileSize,
			},
		).Return(expectedServiceResponse, nil)

		resp, err := fileHandler.CreateMultipartUpload(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedServiceResponse.FileID, resp.FileId)
		assert.Equal(t, expectedServiceResponse.UploadID, resp.UploadId)
		assert.Equal(t, expectedServiceResponse.ObjectKey, resp.ObjectKey)
		assert.Equal(t, expectedServiceResponse.PartSize, resp.PartSize)
		assert.Equal(t, expectedServiceResponse.TotalParts, resp.TotalParts)
	})

	t.Run("failure - validation error (empty file name)", func(t *testing.T) {
		req := &filev1.CreateMultipartUploadRequest{
			FileName:    "", // Invalid
			ContentType: contentType,
			Size:        fileSize,
		}

		resp, err := fileHandler.CreateMultipartUpload(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Equal(t, "validation failed", st.Message())
		assert.Nil(t, resp)
		mockFileService.EXPECT().CreateMultipartUpload(gomock.Any(), gomock.Any()).Times(0)
	})

	t.Run("failure - service returns error", func(t *testing.T) {
		req := &filev1.CreateMultipartUploadRequest{
			FileName:    fileName,
			ContentType: contentType,
			Size:        fileSize,
		}
		serviceErr := errors.New("service multipart upload error")

		mockFileService.EXPECT().CreateMultipartUpload(
			ctx,
			&service.CreateMultipartUploadRequest{
				FileName:    fileName,
				ContentType: contentType,
				Size:        fileSize,
			},
		).Return(nil, serviceErr)

		resp, err := fileHandler.CreateMultipartUpload(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Equal(t, "internal server error", st.Message())
		assert.Nil(t, resp)
	})

	t.Run("failure - service returns specific errs.ErrFileTypeNotAllowed", func(t *testing.T) {
		req := &filev1.CreateMultipartUploadRequest{
			FileName:    fileName,
			ContentType: "application/octet-stream", // Not allowed
			Size:        fileSize,
		}

		mockFileService.EXPECT().CreateMultipartUpload(
			ctx,
			&service.CreateMultipartUploadRequest{
				FileName:    fileName,
				ContentType: "application/octet-stream",
				Size:        fileSize,
			},
		).Return(nil, errs.ErrFileTypeNotAllowed)

		resp, err := fileHandler.CreateMultipartUpload(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), errs.ErrFileTypeNotAllowed.Error())
		assert.Nil(t, resp)
	})
}

func TestFileHandler_CreateMultipartUploadUrl(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileService := mocks.NewMockFileService(ctrl)
	logger := zap.NewNop()

	fileHandler := handler.NewFileHandler(logger, mockFileService)

	ctx := context.Background()
	fileID := "test-file-id"
	uploadID := "test-upload-id"
	objectKey := "test-object-key"
	partNumber := int32(1)

	t.Run("success - valid request", func(t *testing.T) {
		req := &filev1.CreateMultipartUploadUrlRequest{
			FileId:     fileID,
			UploadId:   uploadID,
			ObjectKey:  objectKey,
			PartNumber: partNumber,
		}
		expectedServiceResponse := &service.CreateMultipartUploadUrlResponse{
			UploadURL: "http://presigned.url/part1",
			Headers:   map[string]string{"x-amz-acl": "public-read"},
		}

		mockFileService.EXPECT().CreateMultipartUploadUrl(
			ctx,
			&service.CreateMultipartUploadUrlRequest{
				FileID:     fileID,
				UploadID:   uploadID,
				ObjectKey:  objectKey,
				PartNumber: partNumber,
			},
		).Return(expectedServiceResponse, nil)

		resp, err := fileHandler.CreateMultipartUploadUrl(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedServiceResponse.UploadURL, resp.UploadUrl)
		assert.Equal(t, expectedServiceResponse.Headers, resp.Headers)
	})

	t.Run("failure - validation error (empty file id)", func(t *testing.T) {
		req := &filev1.CreateMultipartUploadUrlRequest{
			FileId:     "", // Invalid
			UploadId:   uploadID,
			ObjectKey:  objectKey,
			PartNumber: partNumber,
		}

		resp, err := fileHandler.CreateMultipartUploadUrl(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Equal(t, "validation failed", st.Message())
		assert.Nil(t, resp)
		mockFileService.EXPECT().CreateMultipartUploadUrl(gomock.Any(), gomock.Any()).Times(0)
	})

	t.Run("failure - service returns error", func(t *testing.T) {
		req := &filev1.CreateMultipartUploadUrlRequest{
			FileId:     fileID,
			UploadId:   uploadID,
			ObjectKey:  objectKey,
			PartNumber: partNumber,
		}
		serviceErr := errors.New("service create multipart upload url error")

		mockFileService.EXPECT().CreateMultipartUploadUrl(
			ctx,
			&service.CreateMultipartUploadUrlRequest{
				FileID:     fileID,
				UploadID:   uploadID,
				ObjectKey:  objectKey,
				PartNumber: partNumber,
			},
		).Return(nil, serviceErr)

		resp, err := fileHandler.CreateMultipartUploadUrl(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Equal(t, "internal server error", st.Message())
		assert.Nil(t, resp)
	})

	t.Run("failure - service returns specific errs.ErrFileNotFound", func(t *testing.T) {
		req := &filev1.CreateMultipartUploadUrlRequest{
			FileId:     fileID,
			UploadId:   uploadID,
			ObjectKey:  objectKey,
			PartNumber: partNumber,
		}

		mockFileService.EXPECT().CreateMultipartUploadUrl(
			ctx,
			&service.CreateMultipartUploadUrlRequest{
				FileID:     fileID,
				UploadID:   uploadID,
				ObjectKey:  objectKey,
				PartNumber: partNumber,
			},
		).Return(nil, errs.ErrFileNotFound)

		resp, err := fileHandler.CreateMultipartUploadUrl(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Contains(t, st.Message(), errs.ErrFileNotFound.Error())
		assert.Nil(t, resp)
	})

	t.Run("failure - service returns specific errs.ErrMultipartUploadMismatch", func(t *testing.T) {
		req := &filev1.CreateMultipartUploadUrlRequest{
			FileId:     fileID,
			UploadId:   uploadID,
			ObjectKey:  objectKey,
			PartNumber: partNumber,
		}

		mockFileService.EXPECT().CreateMultipartUploadUrl(
			ctx,
			&service.CreateMultipartUploadUrlRequest{
				FileID:     fileID,
				UploadID:   uploadID,
				ObjectKey:  objectKey,
				PartNumber: partNumber,
			},
		).Return(nil, errs.ErrMultipartUploadMismatch)

		resp, err := fileHandler.CreateMultipartUploadUrl(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, st.Code())
		assert.Contains(t, st.Message(), errs.ErrMultipartUploadMismatch.Error())
		assert.Nil(t, resp)
	})
}

func TestFileHandler_CompleteMultipartUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileService := mocks.NewMockFileService(ctrl)
	logger := zap.NewNop()

	fileHandler := handler.NewFileHandler(logger, mockFileService)

	ctx := context.Background()
	uploadID := "test-upload-id"
	objectKey := "test-object-key"
	fileID := "test-file-id"

	parts := []*filev1.MultipartPart{
		{PartNumber: 1, Etag: "etag1"},
		{PartNumber: 2, Etag: "etag2"},
	}

	t.Run("success - valid request", func(t *testing.T) {
		req := &filev1.CompleteMultipartUploadRequest{
			UploadId:  uploadID,
			ObjectKey: objectKey,
			Parts:     parts,
		}
		expectedServiceResponse := &service.CompleteMultipartUploadResponse{
			File: &model.File{
				ID:     fileID,
				Status: model.FileStatusCompleted,
			},
		}

		mockFileService.EXPECT().CompleteMultipartUpload(
			ctx,
			gomock.Any(), // Will check the content of this later if needed
		).Return(expectedServiceResponse, nil)

		resp, err := fileHandler.CompleteMultipartUpload(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, fileID, resp.File.Id)
		assert.Equal(t, filev1.FileStatus_FILE_STATUS_COMPLETED, resp.File.Status)
	})

	t.Run("failure - validation error (empty upload id)", func(t *testing.T) {
		req := &filev1.CompleteMultipartUploadRequest{
			UploadId:  "", // Invalid
			ObjectKey: objectKey,
			Parts:     parts,
		}

		resp, err := fileHandler.CompleteMultipartUpload(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Equal(t, "validation failed", st.Message())
		assert.Nil(t, resp)
		mockFileService.EXPECT().CompleteMultipartUpload(gomock.Any(), gomock.Any()).Times(0)
	})

	t.Run("failure - service returns error", func(t *testing.T) {
		req := &filev1.CompleteMultipartUploadRequest{
			UploadId:  uploadID,
			ObjectKey: objectKey,
			Parts:     parts,
		}
		serviceErr := errors.New("service complete multipart upload error")

		mockFileService.EXPECT().CompleteMultipartUpload(
			ctx,
			gomock.Any(),
		).Return(nil, serviceErr)

		resp, err := fileHandler.CompleteMultipartUpload(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Equal(t, "internal server error", st.Message())
		assert.Nil(t, resp)
	})

	t.Run("failure - service returns specific errs.ErrFileNotFound", func(t *testing.T) {
		req := &filev1.CompleteMultipartUploadRequest{
			UploadId:  uploadID,
			ObjectKey: objectKey,
			Parts:     parts,
		}

		mockFileService.EXPECT().CompleteMultipartUpload(
			ctx,
			gomock.Any(),
		).Return(nil, errs.ErrFileNotFound)

		resp, err := fileHandler.CompleteMultipartUpload(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Contains(t, st.Message(), errs.ErrFileNotFound.Error())
		assert.Nil(t, resp)
	})

	t.Run("failure - service returns specific errs.ErrMultipartUploadMismatch", func(t *testing.T) {
		req := &filev1.CompleteMultipartUploadRequest{
			UploadId:  uploadID,
			ObjectKey: objectKey,
			Parts:     parts,
		}

		mockFileService.EXPECT().CompleteMultipartUpload(
			ctx,
			gomock.Any(),
		).Return(nil, errs.ErrMultipartUploadMismatch)

		resp, err := fileHandler.CompleteMultipartUpload(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, st.Code())
		assert.Contains(t, st.Message(), errs.ErrMultipartUploadMismatch.Error())
		assert.Nil(t, resp)
	})
}

func TestFileHandler_AbortMultipartUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileService := mocks.NewMockFileService(ctrl)
	logger := zap.NewNop()

	fileHandler := handler.NewFileHandler(logger, mockFileService)

	ctx := context.Background()
	uploadID := "test-upload-id"
	objectKey := "test-object-key"

	t.Run("success - valid request", func(t *testing.T) {
		req := &filev1.AbortMultipartUploadRequest{
			UploadId:  uploadID,
			ObjectKey: objectKey,
		}

		mockFileService.EXPECT().AbortMultipartUpload(
			ctx,
			&service.AbortMultipartUploadRequest{
				UploadID:  uploadID,
				ObjectKey: objectKey,
			},
		).Return(&service.AbortMultipartUploadResponse{}, nil)

		resp, err := fileHandler.AbortMultipartUpload(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("failure - validation error (empty upload id)", func(t *testing.T) {
		req := &filev1.AbortMultipartUploadRequest{
			UploadId:  "", // Invalid
			ObjectKey: objectKey,
		}

		resp, err := fileHandler.AbortMultipartUpload(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Equal(t, "validation failed", st.Message())
		assert.Nil(t, resp)
		mockFileService.EXPECT().AbortMultipartUpload(gomock.Any(), gomock.Any()).Times(0)
	})

	t.Run("failure - service returns error", func(t *testing.T) {
		req := &filev1.AbortMultipartUploadRequest{
			UploadId:  uploadID,
			ObjectKey: objectKey,
		}
		serviceErr := errors.New("service abort multipart upload error")

		mockFileService.EXPECT().AbortMultipartUpload(
			ctx,
			&service.AbortMultipartUploadRequest{
				UploadID:  uploadID,
				ObjectKey: objectKey,
			},
		).Return(nil, serviceErr)

		resp, err := fileHandler.AbortMultipartUpload(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Equal(t, "internal server error", st.Message())
		assert.Nil(t, resp)
	})

	t.Run("failure - service returns specific errs.ErrFileNotFound", func(t *testing.T) {
		req := &filev1.AbortMultipartUploadRequest{
			UploadId:  uploadID,
			ObjectKey: objectKey,
		}

		mockFileService.EXPECT().AbortMultipartUpload(
			ctx,
			&service.AbortMultipartUploadRequest{
				UploadID:  uploadID,
				ObjectKey: objectKey,
			},
		).Return(nil, errs.ErrFileNotFound)

		resp, err := fileHandler.AbortMultipartUpload(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Contains(t, st.Message(), errs.ErrFileNotFound.Error())
		assert.Nil(t, resp)
	})

	t.Run("failure - service returns specific errs.ErrMultipartUploadMismatch", func(t *testing.T) {
		req := &filev1.AbortMultipartUploadRequest{
			UploadId:  uploadID,
			ObjectKey: objectKey,
		}

		mockFileService.EXPECT().AbortMultipartUpload(
			ctx,
			&service.AbortMultipartUploadRequest{
				UploadID:  uploadID,
				ObjectKey: objectKey,
			},
		).Return(nil, errs.ErrMultipartUploadMismatch)

		resp, err := fileHandler.AbortMultipartUpload(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, st.Code())
		assert.Contains(t, st.Message(), errs.ErrMultipartUploadMismatch.Error())
		assert.Nil(t, resp)
	})
}

func TestFileHandler_GetFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileService := mocks.NewMockFileService(ctrl)
	logger := zap.NewNop()

	fileHandler := handler.NewFileHandler(logger, mockFileService)

	ctx := context.Background()
	fileID := "test-file-id"

	t.Run("success - valid request", func(t *testing.T) {
		req := &filev1.GetFileRequest{FileId: fileID}
		expectedServiceResponse := &service.GetFileResponse{
			File: &model.File{
				ID:          fileID,
				FileName:    "test.jpg",
				ContentType: "image/jpeg",
				Size:        1024,
				Status:      model.FileStatusCompleted,
			},
		}

		mockFileService.EXPECT().GetFile(
			ctx,
			&service.GetFileRequest{FileID: fileID},
		).Return(expectedServiceResponse, nil)

		resp, err := fileHandler.GetFile(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, fileID, resp.File.Id)
		assert.Equal(t, "test.jpg", resp.File.FileName)
	})

	t.Run("failure - validation error (empty file id)", func(t *testing.T) {
		req := &filev1.GetFileRequest{FileId: ""} // Invalid

		resp, err := fileHandler.GetFile(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Equal(t, "validation failed", st.Message())
		assert.Nil(t, resp)
		mockFileService.EXPECT().GetFile(gomock.Any(), gomock.Any()).Times(0)
	})

	t.Run("failure - service returns error", func(t *testing.T) {
		req := &filev1.GetFileRequest{FileId: fileID}
		serviceErr := errors.New("service get file error")

		mockFileService.EXPECT().GetFile(
			ctx,
			&service.GetFileRequest{FileID: fileID},
		).Return(nil, serviceErr)

		resp, err := fileHandler.GetFile(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Equal(t, "internal server error", st.Message())
		assert.Nil(t, resp)
	})

	t.Run("failure - service returns specific errs.ErrFileNotFound", func(t *testing.T) {
		req := &filev1.GetFileRequest{FileId: fileID}

		mockFileService.EXPECT().GetFile(
			ctx,
			&service.GetFileRequest{FileID: fileID},
		).Return(nil, errs.ErrFileNotFound)

		resp, err := fileHandler.GetFile(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Contains(t, st.Message(), errs.ErrFileNotFound.Error())
		assert.Nil(t, resp)
	})
}

func TestFileHandler_CreateDownloadURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileService := mocks.NewMockFileService(ctrl)
	logger := zap.NewNop()

	fileHandler := handler.NewFileHandler(logger, mockFileService)

	ctx := context.Background()
	fileID := "test-file-id"

	t.Run("success - valid request", func(t *testing.T) {
		req := &filev1.CreateDownloadUrlRequest{FileId: fileID}
		expectedServiceResponse := &service.CreateDownloadURLResponse{
			DownloadURL: "http://presigned.url/download",
		}

		mockFileService.EXPECT().CreateDownloadURL(
			ctx,
			&service.CreateDownloadURLRequest{FileID: fileID},
		).Return(expectedServiceResponse, nil)

		resp, err := fileHandler.CreateDownloadUrl(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedServiceResponse.DownloadURL, resp.DownloadUrl)
	})

	t.Run("failure - validation error (empty file id)", func(t *testing.T) {
		req := &filev1.CreateDownloadUrlRequest{FileId: ""} // Invalid

		resp, err := fileHandler.CreateDownloadUrl(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Equal(t, "validation failed", st.Message())
		assert.Nil(t, resp)
		mockFileService.EXPECT().CreateDownloadURL(gomock.Any(), gomock.Any()).Times(0)
	})

	t.Run("failure - service returns error", func(t *testing.T) {
		req := &filev1.CreateDownloadUrlRequest{FileId: fileID}
		serviceErr := errors.New("service create download url error")

		mockFileService.EXPECT().CreateDownloadURL(
			ctx,
			&service.CreateDownloadURLRequest{FileID: fileID},
		).Return(nil, serviceErr)

		resp, err := fileHandler.CreateDownloadUrl(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Equal(t, "internal server error", st.Message())
		assert.Nil(t, resp)
	})

	t.Run("failure - service returns specific errs.ErrFileNotFound", func(t *testing.T) {
		req := &filev1.CreateDownloadUrlRequest{FileId: fileID}

		mockFileService.EXPECT().CreateDownloadURL(
			ctx,
			&service.CreateDownloadURLRequest{FileID: fileID},
		).Return(nil, errs.ErrFileNotFound)

		resp, err := fileHandler.CreateDownloadUrl(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Contains(t, st.Message(), errs.ErrFileNotFound.Error())
		assert.Nil(t, resp)
	})
}

func TestFileHandler_DeleteFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileService := mocks.NewMockFileService(ctrl)
	logger := zap.NewNop()

	fileHandler := handler.NewFileHandler(logger, mockFileService)

	ctx := context.Background()
	fileID := "test-file-id"

	t.Run("success - valid request", func(t *testing.T) {
		req := &filev1.DeleteFileRequest{FileId: fileID}
		expectedServiceResponse := &service.DeleteFileResponse{}

		mockFileService.EXPECT().DeleteFile(
			ctx,
			&service.DeleteFileRequest{FileID: fileID},
		).Return(expectedServiceResponse, nil)

		resp, err := fileHandler.DeleteFile(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, &filev1.DeleteFileResponse{}, resp)
	})

	t.Run("failure - validation error (empty file id)", func(t *testing.T) {
		req := &filev1.DeleteFileRequest{FileId: ""} // Invalid

		resp, err := fileHandler.DeleteFile(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Equal(t, "validation failed", st.Message())
		assert.Nil(t, resp)
		mockFileService.EXPECT().DeleteFile(gomock.Any(), gomock.Any()).Times(0)
	})

	t.Run("failure - service returns error", func(t *testing.T) {
		req := &filev1.DeleteFileRequest{FileId: fileID}
		serviceErr := errors.New("service delete file error")

		mockFileService.EXPECT().DeleteFile(
			ctx,
			&service.DeleteFileRequest{FileID: fileID},
		).Return(nil, serviceErr)

		resp, err := fileHandler.DeleteFile(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Equal(t, "internal server error", st.Message())
		assert.Nil(t, resp)
	})

	t.Run("failure - service returns specific errs.ErrFileNotFound", func(t *testing.T) {
		req := &filev1.DeleteFileRequest{FileId: fileID}

		mockFileService.EXPECT().DeleteFile(
			ctx,
			&service.DeleteFileRequest{FileID: fileID},
		).Return(nil, errs.ErrFileNotFound)

		resp, err := fileHandler.DeleteFile(ctx, req)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Contains(t, st.Message(), errs.ErrFileNotFound.Error())
		assert.Nil(t, resp)
	})
}