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
