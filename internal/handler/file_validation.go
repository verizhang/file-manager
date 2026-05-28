package grpc

import (
	"strings"

	filev1 "github.com/verizhang/file-manager/gen/go/file/v1"
	"github.com/verizhang/file-manager/internal/errs"
)

func validateCreateUploadUrlRequest(
	req *filev1.CreateUploadUrlRequest,
) error {
	violations := []errs.FieldViolation{}

	if strings.TrimSpace(req.GetFileName()) == "" {
		violations = append(violations, errs.FieldViolation{
			Field:       "file_name",
			Description: "file_name is required",
		})
	}

	if strings.TrimSpace(req.GetContentType()) == "" {
		violations = append(violations, errs.FieldViolation{
			Field:       "content_type",
			Description: "content_type is required",
		})
	}

	if req.GetSize() <= 0 {
		violations = append(violations, errs.FieldViolation{
			Field:       "size",
			Description: "size must be greater than 0",
		})
	}

	if len(violations) == 0 {
		return nil
	}

	return errs.NewInvalidArgumentError(
		"validation failed",
		violations,
	)
}

func validateCompleteUploadByFileIdRequest(
	req *filev1.CompleteUploadRequest,
) error {
	violations := []errs.FieldViolation{}

	if strings.TrimSpace(req.GetFileId()) == "" {
		violations = append(violations, errs.FieldViolation{
			Field:       "file_id",
			Description: "file_id is required",
		})
	}

	if len(violations) == 0 {
		return nil
	}

	return errs.NewInvalidArgumentError(
		"validation failed",
		violations,
	)
}

// New validation functions for multipart upload

func validateCreateMultipartUploadRequest(
	req *filev1.CreateMultipartUploadRequest,
) error {
	violations := []errs.FieldViolation{}

	if strings.TrimSpace(req.GetFileName()) == "" {
		violations = append(violations, errs.FieldViolation{
			Field:       "file_name",
			Description: "file_name is required",
		})
	}

	if strings.TrimSpace(req.GetContentType()) == "" {
		violations = append(violations, errs.FieldViolation{
			Field:       "content_type",
			Description: "content_type is required",
		})
	}

	if req.GetSize() <= 0 {
		violations = append(violations, errs.FieldViolation{
			Field:       "size",
			Description: "size must be greater than 0",
		})
	}

	if len(violations) == 0 {
		return nil
	}

	return errs.NewInvalidArgumentError(
		"validation failed",
		violations,
	)
}

func validateCreateMultipartUploadUrlRequest(
	req *filev1.CreateMultipartUploadUrlRequest,
) error {
	violations := []errs.FieldViolation{}

	if strings.TrimSpace(req.GetFileId()) == "" {
		violations = append(violations, errs.FieldViolation{
			Field:       "file_id",
			Description: "file_id is required",
		})
	}

	if strings.TrimSpace(req.GetUploadId()) == "" {
		violations = append(violations, errs.FieldViolation{
			Field:       "upload_id",
			Description: "upload_id is required",
		})
	}

	if strings.TrimSpace(req.GetObjectKey()) == "" {
		violations = append(violations, errs.FieldViolation{
			Field:       "object_key",
			Description: "object_key is required",
		})
	}

	if req.GetPartNumber() <= 0 {
		violations = append(violations, errs.FieldViolation{
			Field:       "part_number",
			Description: "part_number must be greater than 0",
		})
	}

	if len(violations) == 0 {
		return nil
	}

	return errs.NewInvalidArgumentError(
		"validation failed",
		violations,
	)
}

func validateCompleteMultipartUploadRequest(
	req *filev1.CompleteMultipartUploadRequest,
) error {
	violations := []errs.FieldViolation{}

	if strings.TrimSpace(req.GetUploadId()) == "" {
		violations = append(violations, errs.FieldViolation{
			Field:       "upload_id",
			Description: "upload_id is required",
		})
	}

	if strings.TrimSpace(req.GetObjectKey()) == "" {
		violations = append(violations, errs.FieldViolation{
			Field:       "object_key",
			Description: "object_key is required",
		})
	}

	if len(req.GetParts()) == 0 {
		violations = append(violations, errs.FieldViolation{
			Field:       "parts",
			Description: "at least one part is required",
		})
	} else {
		for i, part := range req.GetParts() {
			if part.GetPartNumber() <= 0 {
				violations = append(violations, errs.FieldViolation{
					Field:       "parts",
					Description: "part_number must be greater than 0 for part " + string(rune(i)),
				})
			}
			if strings.TrimSpace(part.GetEtag()) == "" {
				violations = append(violations, errs.FieldViolation{
					Field:       "parts",
					Description: "etag is required for part " + string(rune(i)),
				})
			}
		}
	}

	if len(violations) == 0 {
		return nil
	}

	return errs.NewInvalidArgumentError(
		"validation failed",
		violations,
	)
}

func validateAbortMultipartUploadRequest(
	req *filev1.AbortMultipartUploadRequest,
) error {
	violations := []errs.FieldViolation{}

	if strings.TrimSpace(req.GetUploadId()) == "" {
		violations = append(violations, errs.FieldViolation{
			Field:       "upload_id",
			Description: "upload_id is required",
		})
	}

	if strings.TrimSpace(req.GetObjectKey()) == "" {
		violations = append(violations, errs.FieldViolation{
			Field:       "object_key",
			Description: "object_key is required",
		})
	}

	if len(violations) == 0 {
		return nil
	}

	return errs.NewInvalidArgumentError(
		"validation failed",
		violations,
	)
}