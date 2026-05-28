package errs

import "errors"

var (
	ErrGeneratePresignedURL        = errors.New("failed to generate presigned url")
	ErrCreateFileMetadata          = errors.New("failed to create file metadata")
	ErrFileNotFound                = errors.New("file not found")
	ErrInternal                    = errors.New("internal server error")
	ErrFileNotUploaded             = errors.New("file not uploaded")
	ErrCreateMultipartUpload       = errors.New("failed to create multipart upload")
	ErrCompleteMultipartUpload     = errors.New("failed to complete multipart upload")
	ErrAbortMultipartUpload        = errors.New("failed to abort multipart upload")
	ErrMultipartUploadNotFound     = errors.New("multipart upload not found")
	ErrInvalidMultipartUploadParts = errors.New("invalid multipart upload parts")
)
