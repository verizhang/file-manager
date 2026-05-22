package errs

import "errors"

var (
	ErrGeneratePresignedURL = errors.New("failed to generate presigned url")
	ErrCreateFileMetadata   = errors.New("failed to create file metadata")
	ErrFileNotFound         = errors.New("file not found")
	ErrInternal             = errors.New("internal server error")
	ErrFileNotUploaded      = errors.New("file not uploaded")
)
