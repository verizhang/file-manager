package grpc

import (
	"context"

	filev1 "github.com/verizhang/file-manager/gen/go/file/v1"
	"github.com/verizhang/file-manager/internal/service"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileHandler struct {
	filev1.UnimplementedFileServiceServer
	logger      *zap.Logger
	fileService service.FileService
}

func NewFileHandler(
	logger *zap.Logger,
	fileService service.FileService,

) *FileHandler {
	return &FileHandler{
		logger:      logger,
		fileService: fileService,
	}
}

// =====================================================
// SIMPLE UPLOAD
// =====================================================

func (h *FileHandler) CreateUploadUrl(
	ctx context.Context,
	req *filev1.CreateUploadUrlRequest,
) (*filev1.CreateUploadUrlResponse, error) {

	h.logger.Info("create upload url request received",
		zap.String("file_name", req.GetFileName()),
		zap.String("content_type", req.GetContentType()),
		zap.Int64("size", req.GetSize()),
	)

	response, err := h.fileService.CreateUploadURL(
		ctx,
		req,
	)
	if err != nil {
		h.logger.Error("failed create upload url",
			zap.Error(err),
		)

		return nil, status.Error(
			codes.Internal,
			"failed create upload url",
		)
	}

	return response, nil
}

// =====================================================
// MULTIPART UPLOAD
// =====================================================

func (h *FileHandler) CreateMultipartUpload(
	ctx context.Context,
	req *filev1.CreateMultipartUploadRequest,
) (*filev1.CreateMultipartUploadResponse, error) {

	h.logger.Info("create multipart upload request received",
		zap.String("file_name", req.GetFileName()),
		zap.String("content_type", req.GetContentType()),
		zap.Int64("size", req.GetSize()),
	)

	// TODO(veri):
	// 1. Validate request
	// 2. Calculate total parts
	// 3. Calculate part size
	// 4. Generate object key
	// 5. Create multipart upload session in storage
	// 6. Store upload session metadata
	// 7. Return upload session response

	return nil, status.Error(codes.Unimplemented, "CreateMultipartUpload not implemented")
}

func (h *FileHandler) CreateMultipartUploadUrl(
	ctx context.Context,
	req *filev1.CreateMultipartUploadUrlRequest,
) (*filev1.CreateMultipartUploadUrlResponse, error) {

	h.logger.Info("create multipart upload url request received",
		zap.String("file_id", req.GetFileId()),
		zap.String("upload_id", req.GetUploadId()),
		zap.Int32("part_number", req.GetPartNumber()),
	)

	// TODO(veri):
	// 1. Validate upload session
	// 2. Validate part number
	// 3. Validate upload status
	// 4. Generate presigned multipart upload URL
	// 5. Return upload URL response

	return nil, status.Error(codes.Unimplemented, "CreateMultipartUploadUrl not implemented")
}

func (h *FileHandler) CompleteMultipartUpload(
	ctx context.Context,
	req *filev1.CompleteMultipartUploadRequest,
) (*filev1.CompleteMultipartUploadResponse, error) {

	h.logger.Info("complete multipart upload request received",
		zap.String("upload_id", req.GetUploadId()),
		zap.String("object_key", req.GetObjectKey()),
		zap.Int("parts_count", len(req.GetParts())),
	)

	// TODO(veri):
	// 1. Validate multipart session
	// 2. Validate uploaded parts
	// 3. Complete multipart upload in storage
	// 4. Update file status to completed
	// 5. Trigger async virus scan worker
	// 6. Return file object response

	return nil, status.Error(codes.Unimplemented, "CompleteMultipartUpload not implemented")
}

func (h *FileHandler) AbortMultipartUpload(
	ctx context.Context,
	req *filev1.AbortMultipartUploadRequest,
) (*filev1.AbortMultipartUploadResponse, error) {

	h.logger.Info("abort multipart upload request received",
		zap.String("upload_id", req.GetUploadId()),
		zap.String("object_key", req.GetObjectKey()),
	)

	// TODO(veri):
	// 1. Validate multipart session
	// 2. Abort multipart upload in storage
	// 3. Update file status to failed/aborted
	// 4. Cleanup metadata if necessary

	return nil, status.Error(codes.Unimplemented, "AbortMultipartUpload not implemented")
}

// =====================================================
// FILE ACCESS
// =====================================================

func (h *FileHandler) GetFile(
	ctx context.Context,
	req *filev1.GetFileRequest,
) (*filev1.GetFileResponse, error) {

	h.logger.Info("get file request received",
		zap.String("file_id", req.GetFileId()),
	)

	// TODO(veri):
	// 1. Validate file ID
	// 2. Retrieve file metadata
	// 3. Return file response

	return nil, status.Error(codes.Unimplemented, "GetFile not implemented")
}

func (h *FileHandler) CreateDownloadUrl(
	ctx context.Context,
	req *filev1.CreateDownloadUrlRequest,
) (*filev1.CreateDownloadUrlResponse, error) {

	h.logger.Info("create download url request received",
		zap.String("file_id", req.GetFileId()),
	)

	// TODO(veri):
	// 1. Validate file existence
	// 2. Validate file status
	// 3. Validate virus scan status
	// 4. Generate presigned download URL
	// 5. Return download URL

	return nil, status.Error(codes.Unimplemented, "CreateDownloadUrl not implemented")
}

func (h *FileHandler) CreatePreviewUrl(
	ctx context.Context,
	req *filev1.CreatePreviewUrlRequest,
) (*filev1.CreatePreviewUrlResponse, error) {

	h.logger.Info("create preview url request received",
		zap.String("file_id", req.GetFileId()),
	)

	// TODO(veri):
	// 1. Validate file existence
	// 2. Validate file status
	// 3. Validate previewable content type
	// 4. Validate virus scan status
	// 5. Generate presigned preview URL
	// 6. Return preview URL

	return nil, status.Error(codes.Unimplemented, "CreatePreviewUrl not implemented")
}

// =====================================================
// DELETE
// =====================================================

func (h *FileHandler) DeleteFile(
	ctx context.Context,
	req *filev1.DeleteFileRequest,
) (*filev1.DeleteFileResponse, error) {

	h.logger.Info("delete file request received",
		zap.String("file_id", req.GetFileId()),
	)

	// TODO(veri):
	// 1. Validate file existence
	// 2. Delete object from storage
	// 3. Soft delete metadata
	// 4. Return delete response

	return nil, status.Error(codes.Unimplemented, "DeleteFile not implemented")
}
