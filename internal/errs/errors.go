package errs

import "errors"

var (
	// General
	ErrInternal = errors.New("internal server error")

	// Storage
	ErrGeneratePresignedURL    = errors.New("failed to generate presigned url")
	ErrFileNotUploaded         = errors.New("file not uploaded")
	ErrCreateMultipartUpload   = errors.New("failed to create multipart upload")
	ErrCompleteMultipartUpload = errors.New("failed to complete multipart upload")
	ErrAbortMultipartUpload    = errors.New("failed to abort multipart upload")

	// Service
	ErrMultipartUploadMismatch = errors.New("error multipart upload missmatch")

	// Database
	ErrCreateFileMetadata            = errors.New("failed to create file metadata")
	ErrFileNotFound                  = errors.New("file not found")
	ErrGetFileByID                   = errors.New("failed to get file by id")
	ErrUpdateFileStatus              = errors.New("failed to update file status")
	ErrGetFileByObjectKey            = errors.New("failed to update file by object key")
	ErrUpdateFileStatusAndETag       = errors.New("failed to update file status and ETag")
	UpdateFileStatusAndClearUploadID = errors.New("failed to update file status and clear upload id")
)
