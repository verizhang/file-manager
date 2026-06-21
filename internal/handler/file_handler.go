package grpc

import (
	"context"

	filev1 "github.com/verizhang/file-manager/gen/go/file/v1"
	"github.com/verizhang/file-manager/internal/handler/mapper"
	"github.com/verizhang/file-manager/internal/interceptor"
	"github.com/verizhang/file-manager/internal/service"
	"github.com/verizhang/file-manager/pkg/errs"

	"go.uber.org/zap"
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
	if err := validateCreateUploadUrlRequest(req); err != nil {
		return nil, err
	}

	userID := ctx.Value(interceptor.UserIDContextKey).(string)

	response, err := h.fileService.CreateUploadUrl(
		ctx,
		&service.CreateUploadRequest{
			UserID:      userID,
			FileName:    req.FileName,
			ContentType: req.ContentType,
			Size:        req.Size,
		},
	)
	if err != nil {
		h.logger.Error("failed create upload url",
			zap.Error(err),
		)
		return nil, errs.ToGRPCError(err)
	}

	return &filev1.CreateUploadUrlResponse{
		FileId:    response.File.ID,
		UploadUrl: response.UploadURL,
	}, nil
}

// =====================================================
// COMPLETE UPLOAD
// =====================================================

func (h *FileHandler) CompleteUpload(
	ctx context.Context,
	req *filev1.CompleteUploadRequest,
) (*filev1.CompleteUploadResponse, error) {
	if err := validateCompleteUploadByFileIdRequest(req); err != nil {
		return nil, err
	}

	response, err := h.fileService.CompleteUpload(
		ctx,
		&service.CompleteUploadRequest{
			FileID: req.FileId,
		},
	)
	if err != nil {

		h.logger.Error(
			"failed complete upload",
			zap.Error(err),
		)

		return nil, errs.ToGRPCError(err)
	}

	return &filev1.CompleteUploadResponse{
		File: mapper.ToProtoFile(response.File),
	}, nil
}

// =====================================================
// MULTIPART UPLOAD
// =====================================================

func (h *FileHandler) CreateMultipartUpload(
	ctx context.Context,
	req *filev1.CreateMultipartUploadRequest,
) (*filev1.CreateMultipartUploadResponse, error) {
	if err := validateCreateMultipartUploadRequest(req); err != nil {
		return nil, err
	}

	userID := ctx.Value(interceptor.UserIDContextKey).(string)

	response, err := h.fileService.CreateMultipartUpload(
		ctx,
		&service.CreateMultipartUploadRequest{
			UserID:      userID,
			FileName:    req.FileName,
			ContentType: req.ContentType,
			Size:        req.Size,
		},
	)
	if err != nil {
		h.logger.Error("failed to create multipart upload", zap.Error(err))
		return nil, errs.ToGRPCError(err)
	}

	return &filev1.CreateMultipartUploadResponse{
		FileId:     response.FileID,
		UploadId:   response.UploadID,
		ObjectKey:  response.ObjectKey,
		PartSize:   response.PartSize,
		TotalParts: response.TotalParts,
	}, nil
}

func (h *FileHandler) CreateMultipartUploadUrl(
	ctx context.Context,
	req *filev1.CreateMultipartUploadUrlRequest,
) (*filev1.CreateMultipartUploadUrlResponse, error) {
	if err := validateCreateMultipartUploadUrlRequest(req); err != nil {
		return nil, err
	}

	response, err := h.fileService.CreateMultipartUploadUrl(
		ctx,
		&service.CreateMultipartUploadUrlRequest{
			FileID:     req.FileId,
			UploadID:   req.UploadId,
			ObjectKey:  req.ObjectKey,
			PartNumber: req.PartNumber,
		},
	)
	if err != nil {
		h.logger.Error("failed to create multipart upload URL", zap.Error(err))
		return nil, errs.ToGRPCError(err)
	}

	return &filev1.CreateMultipartUploadUrlResponse{
		UploadUrl: response.UploadURL,
		Headers:   response.Headers,
	}, nil
}

func (h *FileHandler) CompleteMultipartUpload(
	ctx context.Context,
	req *filev1.CompleteMultipartUploadRequest,
) (*filev1.CompleteMultipartUploadResponse, error) {
	if err := validateCompleteMultipartUploadRequest(req); err != nil {
		return nil, err
	}

	serviceParts := make([]service.MultipartPart, len(req.Parts))
	for i, p := range req.Parts {
		serviceParts[i] = service.MultipartPart{
			PartNumber: p.PartNumber,
			ETag:       p.Etag,
		}
	}

	response, err := h.fileService.CompleteMultipartUpload(
		ctx,
		&service.CompleteMultipartUploadRequest{
			UploadID:  req.UploadId,
			ObjectKey: req.ObjectKey,
			Parts:     serviceParts,
		},
	)
	if err != nil {
		h.logger.Error("failed to complete multipart upload", zap.Error(err))
		return nil, errs.ToGRPCError(err)
	}

	return &filev1.CompleteMultipartUploadResponse{
		File: mapper.ToProtoFile(response.File),
	}, nil
}

func (h *FileHandler) AbortMultipartUpload(
	ctx context.Context,
	req *filev1.AbortMultipartUploadRequest,
) (*filev1.AbortMultipartUploadResponse, error) {
	if err := validateAbortMultipartUploadRequest(req); err != nil {
		return nil, err
	}

	_, err := h.fileService.AbortMultipartUpload(
		ctx,
		&service.AbortMultipartUploadRequest{
			UploadID:  req.UploadId,
			ObjectKey: req.ObjectKey,
		},
	)
	if err != nil {
		h.logger.Error("failed to abort multipart upload", zap.Error(err))
		return nil, errs.ToGRPCError(err)
	}

	return &filev1.AbortMultipartUploadResponse{}, nil
}

// =====================================================
// FILE ACCESS
// =====================================================

func (h *FileHandler) GetFile(
	ctx context.Context,
	req *filev1.GetFileRequest,
) (*filev1.GetFileResponse, error) {
	if err := validateGetFileRequest(req); err != nil {
		return nil, err
	}

	response, err := h.fileService.GetFile(
		ctx,
		&service.GetFileRequest{
			FileID: req.GetFileId(),
		},
	)
	if err != nil {
		h.logger.Error("failed to get file",
			zap.Error(err),
		)
		return nil, errs.ToGRPCError(err)
	}

	return &filev1.GetFileResponse{
		File: mapper.ToProtoFile(response.File),
	}, nil
}

func (h *FileHandler) CreateDownloadUrl(
	ctx context.Context,
	req *filev1.CreateDownloadUrlRequest,
) (*filev1.CreateDownloadUrlResponse, error) {
	if err := validateCreateDownloadUrlRequest(req); err != nil {
		return nil, err
	}

	response, err := h.fileService.CreateDownloadURL(
		ctx,
		&service.CreateDownloadURLRequest{
			FileID: req.GetFileId(),
		},
	)
	if err != nil {
		h.logger.Error("failed to create download URL",
			zap.Error(err),
		)
		return nil, errs.ToGRPCError(err)
	}

	return &filev1.CreateDownloadUrlResponse{
		DownloadUrl: response.DownloadURL,
	}, nil
}

// =====================================================
// DELETE
// =====================================================

func (h *FileHandler) DeleteFile(
	ctx context.Context,
	req *filev1.DeleteFileRequest,
) (*filev1.DeleteFileResponse, error) {
	if err := validateDeleteFileRequest(req); err != nil {
		return nil, err
	}

	_, err := h.fileService.DeleteFile(
		ctx,
		&service.DeleteFileRequest{
			FileID: req.GetFileId(),
		},
	)
	if err != nil {
		h.logger.Error("failed to delete file",
			zap.Error(err),
		)
		return nil, errs.ToGRPCError(err)
	}

	return &filev1.DeleteFileResponse{}, nil
}
