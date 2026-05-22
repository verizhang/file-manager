package errs

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ToGRPCError(err error) error {

	switch {

	case errors.Is(err, ErrGeneratePresignedURL):
		return status.Error(
			codes.Internal,
			ErrGeneratePresignedURL.Error(),
		)

	case errors.Is(err, ErrCreateFileMetadata):
		return status.Error(
			codes.Internal,
			ErrCreateFileMetadata.Error(),
		)

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

	default:
		return status.Error(
			codes.Internal,
			ErrInternal.Error(),
		)
	}
}
