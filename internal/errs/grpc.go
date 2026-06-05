package errs

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ToGRPCError(err error) error {

	switch {
	case errors.Is(err, ErrFileNotFound):
		return status.Error(
			codes.NotFound,
			ErrFileNotFound.Error(),
		)

	case errors.Is(err, ErrFileNotUploaded):
		return status.Error(
			codes.NotFound,
			ErrFileNotUploaded.Error(),
		)

	case errors.Is(err, ErrFileTypeNotAllowed):
		return status.Error(
			codes.InvalidArgument,
			ErrFileTypeNotAllowed.Error(),
		)

	case errors.Is(err, ErrFileTooLarge):
		return status.Error(
			codes.InvalidArgument,
			ErrFileTooLarge.Error(),
		)

	case errors.Is(err, ErrMultipartUploadMismatch):
		return status.Error(
			codes.FailedPrecondition,
			ErrMultipartUploadMismatch.Error(),
		)

	default:
		return status.Error(
			codes.Internal,
			ErrInternal.Error(),
		)
	}
}
